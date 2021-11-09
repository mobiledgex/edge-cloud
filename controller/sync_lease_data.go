package main

import (
	"context"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
)

var leaseTimeoutSec int64 = 5
var syncLeaseDataRetry = time.Minute

// This synchronizes lease data with the persistent storage (etcd).
// Data associated with the lease is meant to be deleted after the lease
// expires, if this controller goes away. However, it's also possible
// due to network/cpu congestion, that the lease keepalives fail even
// though both Controller and Etcd are running. If that happens,
// etcd will flush the lease data, and this Controller needs to restore
// it once a new lease can be established.
type SyncLeaseData struct {
	allApis *AllApis
	sync    *Sync
	leaseID int64
	stop    chan struct{}
	cancel  func()
	mux     sync.Mutex
	wg      sync.WaitGroup
}

func NewSyncLeaseData(sy *Sync, allApis *AllApis) *SyncLeaseData {
	syncLeaseData := SyncLeaseData{}
	syncLeaseData.allApis = allApis
	syncLeaseData.sync = sy
	return &syncLeaseData
}

func (s *SyncLeaseData) Start(ctx context.Context) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.stop != nil {
		// already running
		return
	}
	s.stop = make(chan struct{})
	s.wg.Add(1)
	go s.run()
}

func (s *SyncLeaseData) Stop() {
	s.mux.Lock()
	close(s.stop)
	s.mux.Unlock()
	log.DebugLog(log.DebugLevelInfo, "sync lease data stop, waiting")
	s.wg.Wait()
	log.DebugLog(log.DebugLevelInfo, "sync lease data stop done")
	s.mux.Lock()
	s.stop = nil
	s.mux.Unlock()
}

func (s *SyncLeaseData) run() {
	done := false
	errc := make(chan error, 1)

	for !done {
		ctx, cancel := context.WithCancel(context.Background())
		go func(ctx context.Context) {
			err := s.syncData()
			if err == nil {
				// This call blocks until:
				// - underlying keep alive fails
				// - we cancel context
				// Note that if underlying keep alive fails,
				// context is also marked as cancelled (Done).
				err = s.sync.store.KeepAlive(ctx, s.leaseID)
			}
			errc <- err
		}(ctx)
		// wait for sync to be stopped either intentionally or by failure
		select {
		case <-s.stop:
			done = true
		case err := <-errc:
			span := log.StartSpan(log.DebugLevelInfo, "Sync Lease Data recovery")
			ctx := log.ContextWithSpan(context.Background(), span)
			log.SpanLog(ctx, log.DebugLevelInfo, "Sync Lease Data failed", "err", err, "retry-in", syncLeaseDataRetry.String())
			span.Finish()
		}
		cancel()
		// wait before retrying to avoid spinning
		select {
		case <-s.stop:
			done = true
		case <-time.After(syncLeaseDataRetry):
		}
	}
	s.wg.Done()
	log.DebugLog(log.DebugLevelInfo, "sync lease data stopped")
}

func (s *SyncLeaseData) syncData() error {
	span := log.StartSpan(log.DebugLevelInfo, "Sync Lease Data")
	ctx := log.ContextWithSpan(context.Background(), span)
	defer span.Finish()

	// get lease
	leaseID, err := s.sync.store.Grant(ctx, leaseTimeoutSec)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "grant lease failed", "err", err)
		return err
	}
	s.mux.Lock()
	s.leaseID = leaseID
	s.mux.Unlock()

	err = s.allApis.controllerApi.registerController(ctx, leaseID)
	log.SpanLog(ctx, log.DebugLevelInfo, "registered controller", "hostname", cloudcommon.Hostname(), "err", err)
	if err != nil {
		return err
	}
	err = s.allApis.alertApi.syncSourceData(ctx, leaseID)
	log.SpanLog(ctx, log.DebugLevelInfo, "synced alerts", "err", err)
	if err != nil {
		return err
	}
	return nil
}

func (s *SyncLeaseData) LeaseID() int64 {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.leaseID
}

func (s *SyncLeaseData) ControllerAliveLease() int64 {
	return s.LeaseID()
}
