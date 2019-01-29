package main

import (
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
)

var appInstCheckerInterval = 5 * time.Second
var appInstCheckerMaxConcurrent = 40
var appInstStatusTimeout = 2 * time.Second

type AppInstCheck struct {
	cache   edgeproto.AppInstCache
	tagged  map[Tagged]struct{}
	mux     sync.Mutex
	send    *notify.AppInstStatusSend
	checker Checker
}

func (s *AppInstCheck) init(notifyClient *notify.Client) {
	edgeproto.InitAppInstCache(&s.cache)
	s.tagged = make(map[Tagged]struct{})
	// received messages go into the cache, then the cache
	// calls s.updated to trigger the checker
	notifyClient.RegisterRecvAppInstCache(&s.cache)
	s.cache.SetNotifyCb(s.updated)
	// send is used to send statuses back to the controller
	s.send = notify.NewAppInstStatusSend()
	notifyClient.RegisterSend(s.send)
	// checker takes care of running checks in parallel
	s.checker.start(s, appInstCheckerInterval, appInstCheckerMaxConcurrent)
}

func (s *AppInstCheck) updated(key *edgeproto.AppInstKey, old *edgeproto.AppInst) {
	s.mux.Lock()
	s.tagged[*key] = struct{}{}
	s.mux.Unlock()
	s.checker.wakeup()
}

func (s *AppInstCheck) TagAll() {
	s.mux.Lock()
	s.cache.Mux.Lock()
	for k, _ := range s.cache.Objs {
		s.tagged[k] = struct{}{}
	}
	s.cache.Mux.Unlock()
	s.mux.Unlock()
}

func (s *AppInstCheck) GetTagged() map[Tagged]struct{} {
	s.mux.Lock()
	defer s.mux.Unlock()
	tagged := s.tagged
	s.tagged = make(map[Tagged]struct{})
	return tagged
}

func (s *AppInstCheck) CheckOne(t Tagged) {
	key, ok := t.(edgeproto.AppInstKey)
	if !ok {
		panic("invalid type")
	}
	buf := edgeproto.AppInst{}
	if !s.cache.Get(&key, &buf) {
		return
	}
	if buf.State != edgeproto.TrackedState_Ready {
		return
	}
	status := AppInstGetStatus(&buf, appInstStatusTimeout)
	if status != buf.Status {
		// send back new status
		ais := &edgeproto.AppInstStatus{
			Key:    key,
			Status: status,
		}
		s.send.Update(ais)
	}
	// there may be other AppInst checks here as well in the future
}
