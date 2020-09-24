package cli

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mobiledgex/edge-cloud/edgeproto"
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

type WSStreamPayload struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func RunEdgeTurn(req *edgeproto.ExecRequest, exchangeFunc func() (*edgeproto.ExecRequest, error)) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	reply, err := exchangeFunc()
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
