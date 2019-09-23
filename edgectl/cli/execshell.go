package cli

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/util/webrtcutil"
	webrtc "github.com/pion/webrtc/v2"
	"github.com/xtaci/smux"
	"golang.org/x/crypto/ssh/terminal"
)

func WebrtcTunnel(conn net.Listener, dataChan *webrtc.DataChannel, done chan bool, errchan chan error) error {
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
							errchan <- err
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
							errchan <- err
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
		done <- true
	})

	return nil
}

func WebrtcShell(dataChan *webrtc.DataChannel, done chan bool, errchan chan error) error {
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
		done <- true
	})

	return nil
}

func RunWebrtc(req *edgeproto.ExecRequest, exchangeFunc func(offer webrtc.SessionDescription) (*edgeproto.ExecRequest, *webrtc.SessionDescription, error)) error {
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

	timeout := uint16(10000)
	dataChannelOptions := webrtc.DataChannelInit{
		MaxPacketLifeTime: &timeout, // in milliseconds
	}

	dataChan, err := peerConn.CreateDataChannel("data", &dataChannelOptions)
	if err != nil {
		return fmt.Errorf("failed to create data channel, %v", err)
	}

	done := make(chan bool, 1)
	errchan := make(chan error, 1)
	connAddr := ""

	if req.Console {
		conn, err := net.Listen("tcp", "0.0.0.0:0")
		if err != nil {
			return fmt.Errorf("failed to start server, %v", err)
		}
		defer conn.Close()
		connAddr = conn.Addr().String()
		err = WebrtcTunnel(conn, dataChan, done, errchan)
	} else {
		err = WebrtcShell(dataChan, done, errchan)
	}
	if err != nil {
		return err
	}

	offer, err := peerConn.CreateOffer(nil)
	if err != nil {
		return err
	}
	err = peerConn.SetLocalDescription(offer)
	if err != nil {
		return err
	}

	reply, answer, err := exchangeFunc(offer)
	if err != nil {
		return err
	}

	err = peerConn.SetRemoteDescription(*answer)
	if err != nil {
		return err
	}

	if req.Console {
		if connAddr == "" {
			return fmt.Errorf("unable to fetch server address")
		}
		if reply.ConsoleUrl == "" {
			return fmt.Errorf("unable to fetch console URL from webrtc reply")
		}
		urlObj, err := url.Parse(reply.ConsoleUrl)
		if err != nil {
			return fmt.Errorf("unable to parse console url, %s, %v", reply.ConsoleUrl, err)
		}
		util.OpenUrl(strings.Replace(reply.ConsoleUrl, urlObj.Host, connAddr, 1))
	}

	// wait for connection to complete
	var outerr error
	select {
	case <-done:
		err = nil
	case outerr = <-errchan:
		err = outerr
	}

	return err
}
