package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"github.com/stretchr/testify/require"
)

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
}
