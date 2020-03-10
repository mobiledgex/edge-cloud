package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/xtaci/smux"
)

var listenAddr = flag.String("listenAddr", "127.0.0.1:6080", "EdgeTurn listener address")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file")
var tlsKeyFile = flag.String("tlskey", "", "server tls key file")

var sigChan chan os.Signal

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer(*tlsCertFile)
	defer log.FinishTracer()
	span := log.StartSpan(log.DebugLevelInfo, "main")
	ctx := log.ContextWithSpan(context.Background(), span)

	if *listenAddr == "" {
		log.FatalLog("listenAddr is empty")
	}
	var tlsConfig *tls.Config
	if *tlsCertFile != "" && *tlsKeyFile != "" {
		certificate, err := tls.LoadX509KeyPair(*tlsCertFile, *tlsKeyFile)
		if err != nil {
			log.FatalLog("could not load server key pair", "err", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
	}
	var turnConn net.Conn
	if tlsConfig != nil {
		turnConn, err = tls.Listen("tcp", *turnAddr, tlsConfig)
	} else {
		turnConn, err = net.Listen("tcp", *turnAddr)
	}
	if err != nil {
		log.FatalLog("failed to start server", "err", err)
	}
	defer turnConn.Close()
	log.DebugLog(log.DebugLevelApi, "Started EdgeTurn Server")

	for {
		clientConn, err := turnConn.Accept()
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "failed to accept connection", "err", err)
			continue
		}
		go handleConnection(tlsConfig, clientConn)
	}

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	log.SpanLog(ctx, log.DebugLevelInfo, "Ready")
	span.Finish()
	sig := <-sigChan
	fmt.Println(sig)
}

func handleConnection(tlsConfig *tls.Config, clientConn net.Conn) {
	var serverConn net.Conn
	if tlsConfig != nil {
		serverConn, err = tls.Listen("tcp", "0.0.0.0:0", tlsConfig)
	} else {
		serverConn, err = net.Listen("tcp", "0.0.0.0:0")
	}
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "failed to start server", "err", err)
		return
	}
	defer serverConn.Close()

	connAddr := serverConn.Addr().String()
	ports := strings.Split(connAddr, ":")
	connPort := ports[len(ports)-1]
	log.DebugLog(log.DebugLevelApi, "started server", "port", connPort)

	// Send Initial Information about the connection
	sessInfo := util.SessionInfo{
		Port: connPort,
	}
	out, err := json.Marshal(&sessInfo)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "failed to marshal session info", "info", sessInfo, "err", err)
		return
	}
	log.DebugLog(log.DebugLevelApi, "send session info", "info", string(out))
	clientConn.Write(out)

	sess, err := smux.Client(clientConn, nil)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "failed to setup smux server", "err", err)
		return
	}
	for {
		server, err := serverConn.Accept()
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "failed to accept connections", "err", err)
			return
		}

		// Setup proxy
		stream, err := sess.OpenStream()
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "failed to open smux stream", "err", err)
			return
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
