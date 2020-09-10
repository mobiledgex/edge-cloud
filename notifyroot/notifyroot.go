package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"google.golang.org/grpc"
)

var notifyAddr = flag.String("notifyAddr", "127.0.0.1:53001", "Notify listener address")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))

var nodeMgr node.NodeMgr
var sigChan chan os.Signal

func main() {
	nodeMgr.InitFlags()
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer(nodeMgr.TlsCertFile)
	defer log.FinishTracer()
	span := log.StartSpan(log.DebugLevelInfo, "main")
	ctx := log.ContextWithSpan(context.Background(), span)

	err := nodeMgr.Init(ctx, node.NodeTypeNotifyRoot, node.CertIssuerGlobal)
	if err != nil {
		log.FatalLog("Failed to init node", "err", err)
	}

	notifyServer := &notify.ServerMgr{}
	nodeMgr.RegisterServer(notifyServer)
	notifyServer.RegisterServerCb(func(s *grpc.Server) {
		edgeproto.RegisterNodeApiServer(s, &nodeApi)
		edgeproto.RegisterDebugApiServer(s, &debugApi)
	})
	tlsConfig, err := nodeMgr.InternalPki.GetServerTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerGlobal,
		[]node.MatchCA{
			node.AnyRegionalMatchCA(),
			node.GlobalMatchCA(),
		})
	if err != nil {
		log.FatalLog("Failed to get tls config", "err", err)
	}
	notifyServer.Start(nodeMgr.Name(), *notifyAddr, tlsConfig)
	defer notifyServer.Stop()

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	log.SpanLog(ctx, log.DebugLevelInfo, "Ready")
	span.Finish()
	sig := <-sigChan
	fmt.Println(sig)
}
