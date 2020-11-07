package cloudcommon

import (
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var streamCleanupInterval = 10 * time.Minute
var streamExpiration = 10 * time.Minute

type Streamer struct {
	buffer     []interface{}
	mux        sync.Mutex
	subs       map[chan interface{}]struct{}
	State      edgeproto.StreamState
	lastAccess time.Time
}

type Streams map[interface{}]*Streamer

type StreamObj struct {
	streamMap Streams
	mux       sync.Mutex
}

func (sm *StreamObj) Get(in interface{}) *Streamer {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	streamer, found := sm.streamMap[in]
	if found {
		return streamer
	}
	return nil
}

func (sm *StreamObj) Add(in interface{}, streamer *Streamer) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	if sm.streamMap == nil {
		sm.streamMap = Streams{in: streamer}
	} else {
		sm.streamMap[in] = streamer
	}
}

func (sm *StreamObj) Remove(in interface{}, streamer *Streamer) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	if streamerObj, ok := sm.streamMap[in]; ok {
		if streamerObj == streamer {
			delete(sm.streamMap, in)
		}
	}
}

func (sm *StreamObj) SetupCleanupTimer() {
	for {
		select {
		case <-time.After(streamCleanupInterval):
		}
		sm.mux.Lock()
		for obj, streamer := range sm.streamMap {
			streamer.mux.Lock()
			if streamer.State == edgeproto.StreamState_STREAM_START {
				streamer.mux.Unlock()
				continue
			}
			if time.Now().Sub(streamer.lastAccess) >= streamExpiration {
				streamer.subs = nil
				delete(sm.streamMap, obj)
			}
			streamer.mux.Unlock()
		}
		sm.mux.Unlock()
	}
}

func NewStreamer() *Streamer {
	return &Streamer{
		subs:       make(map[chan interface{}]struct{}),
		State:      edgeproto.StreamState_STREAM_START,
		lastAccess: time.Now(),
	}
}

func (s *Streamer) Stop() {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.subs != nil {
		for msgCh := range s.subs {
			close(msgCh)
		}
	}
	s.State = edgeproto.StreamState_STREAM_STOP
	s.lastAccess = time.Now()
}

func (s *Streamer) Subscribe() chan interface{} {
	msgCh := make(chan interface{}, 20)
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.subs == nil {
		// streamer is no longer active
		return nil
	}
	s.subs[msgCh] = struct{}{}
	// Send already streamed msgs to new subscriber
	for _, msg := range s.buffer {
		select {
		case msgCh <- msg:
		default:
		}
	}
	if s.State != edgeproto.StreamState_STREAM_START {
		close(msgCh)
	}
	s.lastAccess = time.Now()
	return msgCh
}

func (s *Streamer) Unsubscribe(msgCh chan interface{}) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.subs == nil {
		// streamer is no longer active
		return
	}
	if _, ok := s.subs[msgCh]; ok {
		delete(s.subs, msgCh)
	}
	s.lastAccess = time.Now()
}

func (s *Streamer) Publish(msg interface{}) {
	// Buffer all the streamed messages till now,
	// so that a newly joined subscriber can get
	// complete list of messages
	s.mux.Lock()
	defer s.mux.Unlock()
	s.buffer = append(s.buffer, msg)
	for msgCh := range s.subs {
		select {
		case msgCh <- msg:
		default:
		}
	}
	s.lastAccess = time.Now()
}
