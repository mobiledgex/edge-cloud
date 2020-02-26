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
var tlsCertFile = flag.String("tls", "", "server tls cert file")

var nodeMgr *node.NodeMgr
var sigChan chan os.Signal

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer(*tlsCertFile)
	defer log.FinishTracer()
	span := log.StartSpan(log.DebugLevelInfo, "main")
	ctx := log.ContextWithSpan(context.Background(), span)

	nodeMgr = node.Init(ctx, "notifyroot")

	notifyServer := &notify.ServerMgr{}
	nodeMgr.RegisterServer(notifyServer)
	notifyServer.RegisterServerCb(func(s *grpc.Server) {
		edgeproto.RegisterNodeApiServer(s, &nodeApi)
	})
	notifyServer.Start(*notifyAddr, *tlsCertFile)
	defer notifyServer.Stop()

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	log.SpanLog(ctx, log.DebugLevelInfo, "Ready")
	span.Finish()
	sig := <-sigChan
	fmt.Println(sig)
}
