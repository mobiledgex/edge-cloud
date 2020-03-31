package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"github.com/stretchr/testify/require"
)

func TestEdgeTurnServer(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi | log.DebugLevelInfo)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	flag.Parse() // set defaults

	*testMode = true

	errChan := make(chan error, 2)
	startServers(ctx, errChan)
	go func() {
		err := <-errChan
		require.Nil(t, err, "start services on EdgeTurn server")
	}()

	// Test session info received for ExecReqShell

	// Connect to EdgeTurn Server
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

	proxyVal := TurnProxy.Get(sessInfo.Token)
	require.NotNil(t, proxyVal, "proxyValue is present")
	port := proxyVal.port
	require.NotEqual(t, "", port, "port is not empty")

	reqHeader := make(http.Header, 1)
	reqHeader.Set("edgetoken", sessInfo.Token)
	dialer := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	ws, _, err := dialer.Dial("wss://127.0.0.1:8443/edgeshell", reqHeader)
	require.Nil(t, err, "client websocket connection to EdgeTurn server")
	defer ws.Close()
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
	n = bytes.Index(msg, []byte{0})
	require.Equal(t, "test msg2", string(msg[:n]), "received message from crm")
}
