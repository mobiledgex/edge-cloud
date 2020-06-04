package cli

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/util/webrtcutil"
	webrtc "github.com/pion/webrtc/v2"
	"github.com/xtaci/smux"
	"golang.org/x/crypto/ssh/terminal"
)

type ConsoleProxyObj struct {
	mux      sync.Mutex
	proxyMap map[string]string
}

var ConsoleProxy = &ConsoleProxyObj{}

func (cp *ConsoleProxyObj) Add(token, port string) {
	cp.mux.Lock()
	defer cp.mux.Unlock()
	if len(cp.proxyMap) == 0 {
		cp.proxyMap = make(map[string]string)
	}
	cp.proxyMap[token] = port
}

func (cp *ConsoleProxyObj) Remove(token string) {
	cp.mux.Lock()
	defer cp.mux.Unlock()
	if _, ok := cp.proxyMap[token]; ok {
		delete(cp.proxyMap, token)
	}
}

func (cp *ConsoleProxyObj) Get(token string) string {
	cp.mux.Lock()
	defer cp.mux.Unlock()
	if out, ok := cp.proxyMap[token]; ok {
		return out
	}
	return ""
}

func WebrtcTunnel(conn net.Listener, dataChan *webrtc.DataChannel, errchan chan error, openurl chan bool) {
	var sess *smux.Session
	dataChan.OnOpen(func() {
		dcconn, err := webrtcutil.WrapDataChannel(dataChan)
		if err != nil {
			errchan <- fmt.Errorf("failed to wrap data channel, %v", err)
			return
		}
		sess, err = smux.Client(dcconn, nil)
		if err != nil {
			errchan <- fmt.Errorf("failed to create smux client, %v", err)
			return
		}

		openurl <- true

		go func() {
			for {
				client, err := conn.Accept()
				if err != nil {
					errchan <- fmt.Errorf("failed to accept connections from %s, %v", conn.Addr().String(), err)
					return
				}
				stream, err := sess.OpenStream()
				if err != nil {
					errchan <- fmt.Errorf("failed to open smux stream, %v", err)
					return
				}
				go func(client net.Conn, stream *smux.Stream) {
					buf := make([]byte, 1500)
					for {
						n, err := stream.Read(buf)
						if err != nil {
							break
						}
						client.Write(buf[:n])
					}
					stream.Close()
					client.Close()
				}(client, stream)

				go func(client net.Conn, stream *smux.Stream) {
					buf := make([]byte, 1500)
					for {
						n, err := client.Read(buf)
						if err != nil {
							break
						}
						stream.Write(buf[:n])
					}
					stream.Close()
					client.Close()
				}(client, stream)
			}
		}()
	})

	dataChan.OnClose(func() {
		if sess != nil {
			sess.Close()
		}
		errchan <- nil
	})
}

// VM Console Handling over edgectl/cli
// - Setup a TCP server on controller to proxy to Infra console URL
//   and make it accessible over end-user's host using edgectl/cli
func SetupLocalConsoleTunnel(consoleUrl string, dataChan *webrtc.DataChannel, errchan chan error, ws *websocket.Conn) {
	urlObj, err := url.Parse(consoleUrl)
	if err != nil {
		errchan <- fmt.Errorf("unable to parse console url, %s, %v", consoleUrl, err)
		return
	}
	queryArgs := urlObj.Query()
	token, ok := queryArgs["token"]
	if !ok || len(token) != 1 {
		errchan <- fmt.Errorf("invalid console url: %s", consoleUrl)
		return
	}
	tlsConfig, err := edgetls.GetLocalTLSConfig()
	if err != nil {
		errchan <- fmt.Errorf("unable to fetch tls local server config, %v", err)
		return
	}

	conn, err := tls.Listen("tcp", "0.0.0.0:0", tlsConfig)
	if err != nil {
		errchan <- fmt.Errorf("failed to start server, %v", err)
		return
	}
	defer func() {
		conn.Close()
	}()

	connAddr := conn.Addr().String()
	ports := strings.Split(connAddr, ":")
	connPort := ports[len(ports)-1]
	connAddr = "127.0.0.1:" + connPort

	tunnelErrChan := make(chan error, 1)
	openurl := make(chan bool, 1)
	WebrtcTunnel(conn, dataChan, tunnelErrChan, openurl)

	// Add token to port map for reverse proxy access
	ConsoleProxy.Add(token[0], connPort)
	defer ConsoleProxy.Remove(token[0])

	for {
		select {
		case <-openurl:
			proxyUrl := strings.Replace(consoleUrl, urlObj.Host, connAddr, 1)
			proxyUrl = strings.Replace(proxyUrl, "http:", "https:", 1)
			if ws != nil {
				wsPayload := WSStreamPayload{
					Code: http.StatusOK,
					Data: proxyUrl,
				}
				err := ws.WriteJSON(wsPayload)
				if err != nil {
					errchan <- fmt.Errorf("failed to write to websocket, %v", err)
					return
				}
				// A close message is sent from client, hence just wait
				// on getting a close message
				_, _, err = ws.ReadMessage()
				if _, ok := err.(*websocket.CloseError); !ok {
					ws.WriteMessage(websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""))
					ws.Close()
					errchan <- nil
					return
				}
				errchan <- err
				return
			}
			dispUrl := strings.Replace(proxyUrl, "127.0.0.1", "<your-host-ip>", 1)
			fmt.Printf("Console URL: %s\n", dispUrl)
			util.OpenUrl(proxyUrl)
			fmt.Println("Press Ctrl-C to exit")
		case err := <-tunnelErrChan:
			errchan <- err
			return
		}
	}
}

type WSStreamPayload struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func WebrtcShellWs(dataChan *webrtc.DataChannel, errchan chan error, ws *websocket.Conn) error {
	wr := webrtcutil.NewDataChanWriter(dataChan)
	dataChan.OnOpen(func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				errchan <- fmt.Errorf("failed to read from websocket, %v", err)
				return
			}
			_, err = wr.Write(msg)
			if err != nil {
				errchan <- fmt.Errorf("failed to write to datachannel, %v", err)
				return
			}
		}
	})
	dataChan.OnMessage(func(msg webrtc.DataChannelMessage) {
		wsPayload := WSStreamPayload{
			Code: http.StatusOK,
			Data: string(msg.Data),
		}
		err := ws.WriteJSON(wsPayload)
		if err != nil {
			errchan <- fmt.Errorf("failed to write to websocket, %v", err)
		}
	})
	dataChan.OnClose(func() {
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws.Close()
		errchan <- nil
	})
	return nil
}

func WebrtcShell(dataChan *webrtc.DataChannel, errchan chan error) error {
	interactive := false
	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		// Set stdin and Stdout to raw
		sinOldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		defer func() {
			terminal.Restore(int(os.Stdin.Fd()), sinOldState)
		}()
		soutOldState, err := terminal.MakeRaw(int(os.Stdout.Fd()))
		if err != nil {
			return err
		}
		defer func() {
			terminal.Restore(int(os.Stdout.Fd()), soutOldState)
		}()
		interactive = true
	}

	wr := webrtcutil.NewDataChanWriter(dataChan)

	dataChan.OnOpen(func() {
		go func() {
			// send stdin to data channel
			io.Copy(wr, os.Stdin)
			// close data channel if input stream is closed
			// in interactive mode. In non-interactive mode,
			// os.Stdin is already closed. Instead we wait
			// until remote end closes data channel.
			// We could also add a timeout for non-interactive mode.
			if interactive {
				dataChan.Close()
			}
		}()
	})
	dataChan.OnMessage(func(msg webrtc.DataChannelMessage) {
		os.Stdout.Write(msg.Data)
	})
	dataChan.OnClose(func() {
		errchan <- nil
	})

	return nil
}

func RunWebrtc(req *edgeproto.ExecRequest, exchangeFunc func(offer *webrtc.SessionDescription) (*edgeproto.ExecRequest, *webrtc.SessionDescription, error), ws *websocket.Conn, setupConsoleTunnel func(consoleUrl string, dataChan *webrtc.DataChannel, errchan chan error, ws *websocket.Conn)) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	// hard code config for now
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{
					"turn:stun.mobiledgex.net:19302",
				},
				Username:       "fake",
				Credential:     "fake",
				CredentialType: webrtc.ICECredentialTypePassword,
			},
		},
	}

	// create a new peer connection
	peerConn, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return fmt.Errorf("failed to establish peer connection, %v", err)
	}

	dataChan, err := peerConn.CreateDataChannel("data", nil)
	if err != nil {
		return fmt.Errorf("failed to create data channel, %v", err)
	}

	offer, err := peerConn.CreateOffer(nil)
	if err != nil {
		return err
	}

	err = peerConn.SetLocalDescription(offer)
	if err != nil {
		return err
	}

	reply, answer, err := exchangeFunc(&offer)
	if err != nil {
		return err
	}

	err = peerConn.SetRemoteDescription(*answer)
	if err != nil {
		return err
	}

	defer dataChan.Close()

	errchan := make(chan error, 1)

	if reply.Console != nil {
		if reply.Console.Url == "" {
			return fmt.Errorf("unable to fetch console URL from webrtc reply")
		}
		_, err := url.Parse(reply.Console.Url)
		if err != nil {
			return fmt.Errorf("unable to parse console url, %s, %v", reply.Console.Url, err)
		}
		go setupConsoleTunnel(reply.Console.Url, dataChan, errchan, ws)
	} else {
		if ws != nil {
			err = WebrtcShellWs(dataChan, errchan, ws)
		} else {
			err = WebrtcShell(dataChan, errchan)
		}
		if err != nil {
			return err
		}
	}

	// wait for connection to complete
	select {
	case <-signalChan:
		return nil
	case err := <-errchan:
		return err
	}
}

func RunEdgeTurn(req *edgeproto.ExecRequest, exchangeFunc func(offer *webrtc.SessionDescription) (*edgeproto.ExecRequest, *webrtc.SessionDescription, error), ws *websocket.Conn) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	reply, _, err := exchangeFunc(nil)
	if err != nil {
		return err
	}

	if reply.AccessUrl == "" {
		return fmt.Errorf("unable to fetch access URL")
	}

	if reply.Console != nil {
		fmt.Println(reply.AccessUrl)
	} else {
		d := websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
			TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
		}
		ws, _, err := d.Dial(reply.AccessUrl, nil)
		if err != nil {
			return err
		}
		defer ws.Close()

		if terminal.IsTerminal(int(os.Stdin.Fd())) {
			// Set stdin and Stdout to raw
			sinOldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
			if err != nil {
				return err
			}
			defer func() {
				terminal.Restore(int(os.Stdin.Fd()), sinOldState)
			}()
			soutOldState, err := terminal.MakeRaw(int(os.Stdout.Fd()))
			if err != nil {
				return err
			}
			defer func() {
				terminal.Restore(int(os.Stdout.Fd()), soutOldState)
			}()
		}

		errChan := make(chan error, 2)
		go func() {
			buf := make([]byte, 1500)
			for {
				n, err := os.Stdin.Read(buf)
				if err != nil {
					if err != io.EOF {
						errChan <- err
					}
					break
				}
				err = ws.WriteMessage(websocket.TextMessage, buf[:n])
				if err != nil {
					if _, ok := err.(*websocket.CloseError); ok {
						errChan <- nil
					} else {
						errChan <- err
					}
					break
				}
			}
		}()
		go func() {
			for {
				_, msg, err := ws.ReadMessage()
				if err != nil {
					if _, ok := err.(*websocket.CloseError); ok {
						errChan <- nil
					} else {
						errChan <- err
					}
					break
				}
				_, err = os.Stdout.Write(msg)
				if err != nil {
					if err == io.EOF {
						errChan <- nil
					} else {
						errChan <- err
					}
					break
				}
			}
		}()
		select {
		case <-signalChan:
		case err = <-errChan:
			return err
		}
	}

	return nil
}
