package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
)

type ControllerApi struct {
	sync  *Sync
	store edgeproto.ControllerStore
	cache edgeproto.ControllerCache
}

var controllerApi = ControllerApi{}

var leaseTimeoutSec int64 = 20
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
func (s *ControllerApi) registerController() error {
	lease, err := s.sync.store.Grant(context.Background(), leaseTimeoutSec)
	if err != nil {
		return err
	}
	controllerAliveLease = lease

	ctrl := edgeproto.Controller{}
	ctrl.Key.Addr = *externalApiAddr
	_, err = s.store.Put(&ctrl, s.sync.syncWait, objstore.WithLease(lease))
	if err != nil {
		return err
	}
	go func() {
		kperr := s.sync.store.KeepAlive(context.Background(), lease)
		if kperr != nil {
			log.FatalLog("KeepAlive failed", "err", kperr)
		}
	}()
	return nil
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
	for _, ctrl := range s.cache.Objs {
		wg.Add(1)
		go func(ctrlAddr string) {
			err := run(arg, ctrlAddr)
			if err != nil {
				mux.Lock()
				if joberr != nil {
					joberr = err
				}
				mux.Unlock()
			}
			wg.Done()
		}(ctrl.Key.Addr)
	}
	s.cache.Mux.Unlock()
	wg.Wait()
	return joberr
}

func ControllerConnect(addr string) (*grpc.ClientConn, error) {
	dialOption, err := tls.GetTLSClientDialOption(addr, *tlsCertFile)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(addr, dialOption)
	if err != nil {
		return nil, fmt.Errorf("Connect to server %s failed: %s", addr, err.Error())
	}
	return conn, nil
}
