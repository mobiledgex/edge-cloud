package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/version"
	"google.golang.org/grpc"
)

type ControllerApi struct {
	sync  *Sync
	store edgeproto.ControllerStore
	cache edgeproto.ControllerCache
}

var controllerApi = ControllerApi{}

var controllerAliveLease int64

func InitControllerApi(sync *Sync) {
	controllerApi.sync = sync
	controllerApi.store = edgeproto.NewControllerStore(sync.store)
	edgeproto.InitControllerCache(&controllerApi.cache)
	sync.RegisterCache(&controllerApi.cache)
}

// register controller puts this controller into the etcd database
// with a lease and keepalive, such that if this controller
// is shut-down/disappears, etcd will automatically remove it
// from the database after the ttl time.
// We use this mechanism to keep track of the controllers that are online.
// Note that the calls to etcd will block if etcd is not reachable.
func (s *ControllerApi) registerController(ctx context.Context, lease int64) error {
	ctrl := edgeproto.Controller{}
	ctrl.Key.Addr = *externalApiAddr
	ctrl.BuildMaster = version.BuildMaster
	ctrl.BuildHead = version.BuildHead
	ctrl.BuildAuthor = version.BuildAuthor
	ctrl.Hostname = cloudcommon.Hostname()
	_, err := s.store.Put(ctx, &ctrl, s.sync.syncWait, objstore.WithLease(lease))
	return err
}

func (s *ControllerApi) ShowController(in *edgeproto.Controller, cb edgeproto.ControllerApi_ShowControllerServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Controller) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

// RunJobs spawns a thread per controller to run the passed in
// function. RunJobs blocks until all threads are done.
func (s *ControllerApi) RunJobs(run func(arg interface{}, addr string) error, arg interface{}) error {
	var joberr error
	var mux sync.Mutex

	wg := sync.WaitGroup{}
	s.cache.Mux.Lock()
	for _, data := range s.cache.Objs {
		ctrl := data.Obj
		wg.Add(1)
		go func(ctrlAddr string) {
			err := run(arg, ctrlAddr)
			if err != nil {
				mux.Lock()
				if err != nil {
					joberr = err
				}
				mux.Unlock()
				log.DebugLog(log.DebugLevelApi, "run job failed", "addr", ctrlAddr, "err", err)
			}
			wg.Done()
		}(ctrl.Key.Addr)
	}
	s.cache.Mux.Unlock()
	wg.Wait()
	return joberr
}

func ControllerConnect(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	if ip := net.ParseIP(host); ip != nil {
		// This is an IP address. Within kubernetes,
		// controllers will need to connect to each other via
		// IP address, which will not be a SAN defined on the cert.
		// So set the hostname(SNI) on the TLS query to a valid SAN.
		host = nodeMgr.CommonName()
	}
	tlsConfig, err := nodeMgr.InternalPki.GetClientTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerRegional,
		[]node.MatchCA{node.SameRegionalMatchCA()},
		node.WithTlsServerName(host))
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(addr,
		tls.GetGrpcDialOption(tlsConfig),
		grpc.WithUnaryInterceptor(log.UnaryClientTraceGrpc),
		grpc.WithStreamInterceptor(log.StreamClientTraceGrpc),
	)
	if err != nil {
		return nil, fmt.Errorf("Connect to server %s failed: %s", addr, err.Error())
	}
	return conn, nil
}

func notifyRootConnect(ctx context.Context, notifyAddrs string) (*grpc.ClientConn, error) {
	if notifyAddrs == "" {
		return nil, fmt.Errorf("No parent notify address specified, cannot connect to notify root")
	}
	addrs := strings.Split(notifyAddrs, ",")
	tlsConfig, err := nodeMgr.InternalPki.GetClientTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerRegional,
		[]node.MatchCA{node.GlobalMatchCA()},
		node.WithTlsServerName(addrs[0]))
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(addrs[0],
		tls.GetGrpcDialOption(tlsConfig),
		grpc.WithUnaryInterceptor(log.UnaryClientTraceGrpc),
		grpc.WithStreamInterceptor(log.StreamClientTraceGrpc),
	)
	if err != nil {
		return nil, fmt.Errorf("Connect to server %s failed: %s", addrs[0], err.Error())
	}
	return conn, nil
}
