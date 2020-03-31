package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util/proxyutil"
	"github.com/segmentio/ksuid"
)

var listenAddr = flag.String("listenAddr", "127.0.0.1:6080", "EdgeTurn listener address")
var proxyAddr = flag.String("proxyAddr", "127.0.0.1:8443", "EdgeTurn Proxy Address")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file")
var testMode = flag.Bool("testMode", false, "Run EdgeTurn in test mode")

const (
	ShellConnTimeout   = 5 * time.Minute
	ConsoleConnTimeout = 20 * time.Minute
)

type ProxyValue struct {
	port    string
	initURL *url.URL
}

type TurnProxyObj struct {
	mux      sync.Mutex
	proxyMap map[string]*ProxyValue
}

func (cp *TurnProxyObj) Add(token string, proxyVal *ProxyValue) {
	if proxyVal == nil {
		return
	}

	cp.mux.Lock()
	defer cp.mux.Unlock()
	if len(cp.proxyMap) == 0 {
		cp.proxyMap = make(map[string]*ProxyValue)
	}
	cp.proxyMap[token] = proxyVal
}

func (cp *TurnProxyObj) Remove(token string) {
	cp.mux.Lock()
	defer cp.mux.Unlock()
	if _, ok := cp.proxyMap[token]; ok {
		delete(cp.proxyMap, token)
	}
}

func (cp *TurnProxyObj) Get(token string) *ProxyValue {
	cp.mux.Lock()
	defer cp.mux.Unlock()
	if out, ok := cp.proxyMap[token]; ok {
		return out
	}
	return nil
}

var (
	sigChan   chan os.Signal
	TurnProxy = &TurnProxyObj{}
)

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer(*tlsCertFile)
	defer log.FinishTracer()

	span := log.StartSpan(log.DebugLevelInfo, "main")
	ctx := log.ContextWithSpan(context.Background(), span)
	defer span.Finish()

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	errChan := make(chan error, 2)

	startServers(ctx, errChan)

	log.SpanLog(ctx, log.DebugLevelInfo, "Ready")

	select {
	case err := <-errChan:
		if err != nil {
			log.FatalLog(err.Error())
		}
	case sig := <-sigChan:
		fmt.Println(sig)
	}

}

func startServers(ctx context.Context, errChan chan error) {
	started := make(chan bool, 2)
	go func() {
		if *listenAddr == "" {
			log.FatalLog("listenAddr is empty")
		}
		err := setupTurnServer(ctx, started)
		if err != nil {
			errChan <- err
		}
	}()
	<-started
	go func() {
		if *proxyAddr == "" {
			log.FatalLog("proxyAddr is empty")
		}
		err := setupProxyServer(ctx, started)
		if err != nil {
			errChan <- err
		}
	}()
	<-started
}

func setupTurnServer(ctx context.Context, started chan bool) error {
	mutualAuth := true
	if *testMode {
		mutualAuth = false
	}
	tlsConfig, err := edgetls.GetTLSServerConfig(*tlsCertFile, mutualAuth)
	if err != nil {
		return err
	}
	if tlsConfig == nil {
		tlsConfig, err = edgetls.GetLocalTLSConfig()
		if err != nil {
			return fmt.Errorf("unable to fetch tls local server config, %v", err)
		}
	}
	turnConn, err := tls.Listen("tcp", *listenAddr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to start server, %v", err)
	}
	defer turnConn.Close()

	log.SpanLog(ctx, log.DebugLevelInfo, "Started EdgeTurn Server")
	started <- true

	for {
		crmConn, err := turnConn.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection, %v", err)
		}
		go handleConnection(ctx, tlsConfig, crmConn)
	}
}

func handleConnection(ctx context.Context, tlsConfig *tls.Config, crmConn net.Conn) {
	// Since this port is not exposed by external proxy, use local certs here
	tlsConfig, err := edgetls.GetLocalTLSConfig()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "unable to fetch tls local server config", "err", err)
		return
	}
	serverListener, err := tls.Listen("tcp", "0.0.0.0:0", tlsConfig)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "failed to start server", "err", err)
		return
	}
	defer serverListener.Close()
	defer crmConn.Close()

	connAddr := serverListener.Addr().String()
	addrParts := strings.Split(connAddr, ":")
	connPort := addrParts[len(addrParts)-1]
	log.SpanLog(ctx, log.DebugLevelInfo, "started server on:", "addr", connAddr)

	// Fetch exec req info
	var execReqInfo cloudcommon.ExecReqInfo
	d := json.NewDecoder(crmConn)
	err = d.Decode(&execReqInfo)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "failed to decode execreq info", "err", err)
		return
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Recieved execreq info", "info", execReqInfo)

	// Generate session token
	tokObj := ksuid.New()
	token := tokObj.String()
	proxyVal := &ProxyValue{
		port:    connPort,
		initURL: execReqInfo.InitURL,
	}
	TurnProxy.Add(token, proxyVal)
	defer func() {
		TurnProxy.Remove(token)
	}()

	// Send Initial Information about the connection
	sessInfo := cloudcommon.SessionInfo{
		Token: token,
	}
	out, err := json.Marshal(&sessInfo)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "failed to marshal session info", "info", sessInfo, "err", err)
		return
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "send session info", "info", string(out))
	crmConn.Write(out)

	errChan := make(chan error)
	switch execReqInfo.Type {
	case cloudcommon.ExecReqShell:
		var serverConn net.Conn
		go func() {
			serverConn, err = serverListener.Accept()
			errChan <- err
		}()
		select {
		case err = <-errChan:
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "server accept failed", "err", err)
				return
			}
		case <-time.After(ShellConnTimeout):
			log.SpanLog(ctx, log.DebugLevelInfo, "timeout waiting for server to accept connection")
			if serverConn != nil {
				serverConn.Close()
			}
			return
		}
		defer serverConn.Close()
		go io.Copy(crmConn, serverConn)
		io.Copy(serverConn, crmConn)
	case cloudcommon.ExecReqConsole:
		go func() {
			errChan <- proxyutil.ProxyMuxClient(serverListener, crmConn)
		}()
		select {
		case err = <-errChan:
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "failed to setup proxy mux client", "err", err)
				return
			}
		case <-time.After(ConsoleConnTimeout):
			log.SpanLog(ctx, log.DebugLevelInfo, "closing console connection, user must reconnect with new token")
		}
	default:
		log.SpanLog(ctx, log.DebugLevelInfo, "Unknown execreq type", "type", execReqInfo.Type)
		return
	}
}

func setupProxyServer(ctx context.Context, started chan bool) error {
	director := func(req *http.Request) {
		token := ""
		queryArgs := req.URL.Query()
		tokenVals, ok := queryArgs["token"]
		copyURL := false
		if !ok || len(tokenVals) != 1 {
			// try token from cookies
			for _, cookie := range req.Cookies() {
				if cookie.Name == "edgetoken" {
					token = cookie.Value
					break
				}
			}
		} else {
			token = tokenVals[0]
			copyURL = true
		}
		proxyVal := TurnProxy.Get(token)
		if proxyVal == nil || proxyVal.port == "" || proxyVal.initURL == nil {
			req.Close = true
			return
		}
		if copyURL {
			req.URL = proxyVal.initURL
		}
		req.URL.Scheme = "https"
		req.URL.Host = "127.0.0.1:" + proxyVal.port
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	proxy := &httputil.ReverseProxy{Director: director}

	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	http.HandleFunc("/edgeconsole", func(w http.ResponseWriter, r *http.Request) {
		queryArgs := r.URL.Query()
		tokenVals, ok := queryArgs["token"]
		if ok && len(tokenVals) == 1 {
			token := tokenVals[0]
			expire := time.Now().Add(10 * time.Minute)
			cookie := http.Cookie{
				Name:    "edgetoken",
				Value:   tokenVals[0],
				Expires: expire,
			}
			http.SetCookie(w, &cookie)
			log.SpanLog(ctx, log.DebugLevelInfo, "setup console proxy cookies", "url", r.URL, "token", token)
		}
		proxy.ServeHTTP(w, r)
	})

	var upgrader = websocket.Upgrader{} // use default options
	http.HandleFunc("/edgeshell", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "failed to upgrade to websocket", "err", err)
			return
		}
		token := r.Header.Get("edgetoken")
		log.SpanLog(ctx, log.DebugLevelInfo, "Found edgetoken", "token", token)
		if token == "" {
			log.SpanLog(ctx, log.DebugLevelInfo, "no edgetoken found")
			r.Close = true
			return
		}
		proxyVal := TurnProxy.Get(token)
		port := ""
		if proxyVal != nil && proxyVal.port != "" {
			port = proxyVal.port
		}
		log.SpanLog(ctx, log.DebugLevelInfo, "Found shell port", "port", port)
		if port == "" {
			log.SpanLog(ctx, log.DebugLevelInfo, "No port found for edgetoken", "token", token)
			r.Close = true
			return
		}
		target := "127.0.0.1:" + port
		proxyConn, err := tls.Dial("tcp", target, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Error dialing backend", "target", target, "err", err)
			return
		}
		defer proxyConn.Close()
		defer c.Close()
		go func() {
			for {
				_, msg, err := c.ReadMessage()
				if err != nil {
					if err != io.EOF {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to read from websocket", "err", err)
					}
					break
				}
				_, err = proxyConn.Write(msg)
				if err != nil {
					if err != io.EOF {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to write to proxyConn", "err", err)
					}
				}
			}
		}()
		for {
			buf := make([]byte, 1500)
			_, err = proxyConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.SpanLog(ctx, log.DebugLevelInfo, "failed to read from proxyConn", "err", err)
				}
				break
			}

			err = c.WriteMessage(websocket.TextMessage, buf)
			if err != nil {
				if err != io.EOF {
					log.SpanLog(ctx, log.DebugLevelInfo, "failed to write to websocket", "err", err)
				}
				break
			}
		}
	})

	log.SpanLog(ctx, log.DebugLevelInfo, "Starting EdgeTurn Proxy Server")
	started <- true

	var err error
	if *testMode {
		// In test mode, setup HTTP server with TLS
		tlsConfig, err := edgetls.GetLocalTLSConfig()
		if err != nil {
			return fmt.Errorf("unable to fetch tls local server config, %v", err)
		}
		server := &http.Server{
			Addr:      *proxyAddr,
			Handler:   nil,
			TLSConfig: tlsConfig,
		}
		err = server.ListenAndServeTLS("", "")
	} else {
		// Certs will be provided by LB
		err = http.ListenAndServe(*proxyAddr, nil)
	}
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("Failed to start console proxy server, %v", err)
	}
	return nil
}
