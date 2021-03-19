package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"github.com/stretchr/testify/require"
	"github.com/xtaci/smux"
)

func setupConsoleStream(sess *smux.Session, consoleUrlHost string, isTLS bool) error {
	for {
		stream, err := sess.AcceptStream()
		if err != nil {
			if err.Error() != io.ErrClosedPipe.Error() {
				return fmt.Errorf("failed to setup smux acceptstream, %v", err)
			}
			return nil
		}
		var server net.Conn
		if isTLS {
			server, err = tls.Dial("tcp", consoleUrlHost, &tls.Config{
				InsecureSkipVerify: true,
			})
			if err != nil {
				return fmt.Errorf("failed to get console, %v", err)
			}
		} else {
			server, err = net.Dial("tcp", consoleUrlHost)
			if err != nil {
				return fmt.Errorf("failed to get console, %v", err)
			}
		}
		go func(server net.Conn, stream *smux.Stream) {
			buf := make([]byte, 1500)
			for {
				n, err := stream.Read(buf)
				if err != nil {
					break
				}
				server.Write(buf[:n])
			}
			stream.Close()
			server.Close()
		}(server, stream)
		go func(server net.Conn, stream *smux.Stream) {
			buf := make([]byte, 1500)
			for {
				n, err := server.Read(buf)
				if err != nil {
					break
				}
				stream.Write(buf[:n])
			}
			stream.Close()
			server.Close()
		}(server, stream)
	}
}

func TestEdgeTurnServer(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi | log.DebugLevelInfo)
	log.InitTracer(nil)
	defer log.FinishTracer()
	flag.Parse() // set defaults

	*testMode = true

	started := make(chan bool)
	go func() {
		err := setupTurnServer(started)
		if err != nil {
			log.FatalLog(err.Error())
		}
	}()
	<-started

	go func() {
		err := setupProxyServer(started)
		if err != nil {
			log.FatalLog(err.Error())
		}
	}()
	<-started

	// Test session info received for ExecReqShell
	// CRM connection to EdgeTurn
	tlsConfig, err := edgetls.GetLocalTLSConfig()
	require.Nil(t, err, "get local tls config")
	turnConn, err := tls.Dial("tcp", "127.0.0.1:6080", tlsConfig)
	require.Nil(t, err, "connect to EdgeTurn server")
	defer turnConn.Close()

	// Send ExecReqInfo to EdgeTurn Server
	execReqInfo := cloudcommon.ExecReqInfo{
		Type: cloudcommon.ExecReqShell,
	}
	out, err := json.Marshal(&execReqInfo)
	require.Nil(t, err, "marshal ExecReqInfo")
	_, err = turnConn.Write(out)
	require.Nil(t, err, "send ExecReqInfo to EdgeTurn server")

	// EdgeTurn Server should reply with SessionInfo
	var sessInfo cloudcommon.SessionInfo
	d := json.NewDecoder(turnConn)
	err = d.Decode(&sessInfo)
	require.Nil(t, err, "decode session info from EdgeTurn server")
	require.NotEqual(t, "", sessInfo.Token, "token is not empty")
	require.Equal(t, "8443", sessInfo.AccessPort, "accessport is set to default value")

	proxyVal := TurnProxy.Get(sessInfo.Token)
	require.NotNil(t, proxyVal, "proxyValue is present, hence not nil")
	require.NotNil(t, proxyVal.CrmConn, "crm connection is not nil")

	// Client connection to EdgeTurn
	dialer := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	ws, _, err := dialer.Dial("wss://127.0.0.1:8443/edgeshell?edgetoken="+sessInfo.Token, nil)
	require.Nil(t, err, "client websocket connection to EdgeTurn server")
	err = ws.WriteMessage(websocket.TextMessage, []byte("test msg1"))
	require.Nil(t, err, "client write message to EdgeTurn server")
	buf := make([]byte, 50)
	n, err := turnConn.Read(buf)
	require.Nil(t, err, "read from EdgeTurn connection")
	require.Equal(t, "test msg1", string(buf[:n]), "received message from client")

	_, err = turnConn.Write([]byte("test msg2"))
	require.Nil(t, err, "write to EdgeTurn connection")

	_, msg, err := ws.ReadMessage()
	require.Nil(t, err, "client read message from EdgeTurn server")
	require.Equal(t, "test msg2", string(msg), "received message from crm")

	// Client closes connection, this should cleanup connection
	// from EdgeTurn server side as well
	ws.Close()
	time.Sleep(1 * time.Second)

	proxyVal = TurnProxy.Get(sessInfo.Token)
	require.Nil(t, proxyVal, "proxyValue should not exist as client exited")

	// Test edge console
	// =================
	isTLS := true
	testEdgeTurnConsole(t, isTLS)
	testEdgeTurnConsole(t, !isTLS)
}

func testEdgeTurnConsole(t *testing.T, isTLS bool) {
	// Start local console server
	var consoleServer *httptest.Server
	if isTLS {
		consoleServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proto, found := r.Header["X-Forwarded-Proto"]
			require.True(t, found, "found x-forwarded-proto header")
			require.Equal(t, proto[0], "https")
			fmt.Fprintln(w, "Console Content")
		}))
	} else {
		consoleServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proto, found := r.Header["X-Forwarded-Proto"]
			require.True(t, found, "found x-forwarded-proto header")
			require.Equal(t, proto[0], "http")
			fmt.Fprintln(w, "Console Content")
		}))
	}
	require.NotNil(t, consoleServer, "start console server")
	consoleURL := consoleServer.URL + "?token=xyz"
	initURL, err := url.Parse(consoleURL)
	require.Nil(t, err)
	// Test session info received for ExecReqShell
	// CRM connection to EdgeTurn
	tlsConfig, err := edgetls.GetLocalTLSConfig()
	require.Nil(t, err, "get local tls config")
	turnConn1, err := tls.Dial("tcp", "127.0.0.1:6080", tlsConfig)
	require.Nil(t, err, "connect to EdgeTurn server")
	defer turnConn1.Close()
	// Send ExecReqInfo to EdgeTurn Server
	execReqInfo := cloudcommon.ExecReqInfo{
		Type:    cloudcommon.ExecReqConsole,
		InitURL: initURL,
	}
	out, err := json.Marshal(&execReqInfo)
	require.Nil(t, err, "marshal ExecReqInfo")
	_, err = turnConn1.Write(out)
	require.Nil(t, err, "send ExecReqInfo to EdgeTurn server")

	// EdgeTurn Server should reply with SessionInfo
	var sessInfo cloudcommon.SessionInfo
	d := json.NewDecoder(turnConn1)
	err = d.Decode(&sessInfo)
	require.Nil(t, err, "decode session info from EdgeTurn server")
	require.NotEqual(t, "", sessInfo.Token, "token is not empty")
	require.Equal(t, "8443", sessInfo.AccessPort, "accessport is set to default value")

	proxyVal := TurnProxy.Get(sessInfo.Token)
	require.NotNil(t, proxyVal, "proxyValue is present, hence not nil")
	require.NotNil(t, proxyVal.CrmConn, "crm connection is not nil")

	// setup SMUX connection to console server
	sess, err := smux.Server(turnConn1, nil)
	require.Nil(t, err, "setup smux server")
	go setupConsoleStream(sess, initURL.Host, isTLS)

	contents, err := util.ReadConsoleURL("https://127.0.0.1:8443/edgeconsole?edgetoken="+sessInfo.Token, nil)
	require.Nil(t, err)
	require.Equal(t, string(contents), "Console Content\n")
}
