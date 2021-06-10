package notify

// Sendrecv handles sending and receiving data streams.
// While the initial connection and negotiation between client and server
// is asymmetric, after that send and recv streams behave the
// same regardless if the node is a client or server.
// Sendrecv code handles this common send/recv logic.

import (
	"context"
	fmt "fmt"
	"sync"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

// NotifySend is implemented by auto-generated code. That code
// is specific to the given object, but the interface is generic.
// The sendrecv code uses the generic interface to treat all
// objects equally.
type NotifySend interface {
	// Set the SendRecv owner
	SetSendRecv(s *SendRecv)
	// Get the proto message name
	GetMessageName() string
	// Get the object name (no package name)
	GetName() string
	// Get send count for object
	GetSendCount() uint64
	// Send the data
	Send(stream StreamNotify, buf *edgeproto.Notice, peerAddr string) error
	// Return true if there are keys to send, prepares keys to send
	PrepData() bool
	// Queue all cached data for send
	UpdateAll(ctx context.Context)
}

// NotifyRecv is implemented by auto-generated code. The same
// comment as for NotifySend applies here as well.
type NotifyRecv interface {
	// Set the SendRecv owner
	SetSendRecv(s *SendRecv)
	// Get the proto message name
	GetMessageName() string
	// Get the object name (no package name)
	GetName() string
	// Get recv count for object
	GetRecvCount() uint64
	// Recieve the data
	Recv(ctx context.Context, notice *edgeproto.Notice, notifyId int64, peerAddr string)
	// Start receiving a send all
	RecvAllStart()
	// End receiving a send all
	RecvAllEnd(ctx context.Context, cleanup Cleanup)
}

type SendAllRecv interface {
	// Called when SendAllStart is received
	RecvAllStart()
	// Called when SendAllEnd is received
	RecvAllEnd(ctx context.Context)
}

type StreamNotify interface {
	Send(*edgeproto.Notice) error
	Recv() (*edgeproto.Notice, error)
	grpc.Stream
}

type Stats struct {
	Tries           uint64
	Connects        uint64
	NegotiateErrors uint64
	SendAll         uint64
	Send            uint64
	Recv            uint64
	RecvErrors      uint64
	SendErrors      uint64
	MarshalErrors   uint64
	UnmarshalErrors uint64
	ObjSend         map[string]uint64
	ObjRecv         map[string]uint64
}

// There are two ways to clean up stale cache entries.
// On Servers, which can handle multiple clients sending
// it the same types of objects, we delete objects when a
// client disconnects (Flush). Only objects associated
// with the disconnected client are flushed.
// On clients, which only connect to one server, and want
// to keep a cache of what the server sent even when
// disconnected, we maintain the cache after disconnect,
// but Prune stale entries after reconnect and receiving
// the SendAllEnd message. All entries not sent by the
// server are pruned (deleted).
type Cleanup int

const (
	CleanupPrune Cleanup = iota
	CleanupFlush
)

type SendRecv struct {
	cliserv            string // client or server
	peerAddr           string
	peer               string
	sendlist           []NotifySend
	recvmap            map[string]NotifyRecv
	started            bool
	done               bool
	localWanted        []string
	remoteWanted       map[string]struct{}
	filterCloudletKeys bool
	cloudletKeys       map[edgeproto.CloudletKey]struct{}
	cloudletReady      bool
	appSend            *AppSend
	cloudletSend       *CloudletSend
	clusterInstSend    *ClusterInstSend
	appInstSend        *AppInstSend
	vmPoolSend         *VMPoolSend
	gpuDriverSend      *GPUDriverSend
	TrustPolicySend    *TrustPolicySend
	sendRunning        chan struct{}
	recvRunning        chan struct{}
	signal             chan bool
	stats              Stats
	mux                sync.Mutex
	sendAllEnd         bool
	manualSendAllEnd   bool
	sendAllRecvHandler SendAllRecv
}

func (s *SendRecv) init(cliserv string) {
	s.cliserv = cliserv
	s.sendlist = make([]NotifySend, 0)
	s.recvmap = make(map[string]NotifyRecv)
	s.localWanted = []string{}
	s.remoteWanted = make(map[string]struct{})
	s.cloudletKeys = make(map[edgeproto.CloudletKey]struct{})
	s.signal = make(chan bool, 1)
}

func (s *SendRecv) registerSend(send NotifySend) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.started {
		log.FatalLog("Must register before starting")
	}
	s.sendlist = append(s.sendlist, send)
	send.SetSendRecv(s)
	// track some specific sends for cloudlet key filtering
	switch v := send.(type) {
	case *AppSend:
		s.appSend = v
	case *VMPoolSend:
		s.vmPoolSend = v
	case *GPUDriverSend:
		s.gpuDriverSend = v
	case *CloudletSend:
		s.cloudletSend = v
	case *ClusterInstSend:
		s.clusterInstSend = v
	case *AppInstSend:
		s.appInstSend = v
	case *TrustPolicySend:
		s.TrustPolicySend = v
	}
}

func (s *SendRecv) registerRecv(recv NotifyRecv) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.started {
		log.FatalLog("Must register before starting")
	}
	s.recvmap[recv.GetMessageName()] = recv
	s.localWanted = append(s.localWanted, recv.GetMessageName())
	recv.SetSendRecv(s)
}

func (s *SendRecv) registerSendAllRecv(handler SendAllRecv) {
	s.sendAllRecvHandler = handler
}

func (s *SendRecv) setRemoteWanted(names []string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	for _, name := range names {
		s.remoteWanted[name] = struct{}{}
	}
}

func (s *SendRecv) triggerSendAllEnd() {
	s.mux.Lock()
	defer s.mux.Unlock()
	// we only allow sendAllEnd to be triggered once
	if s.manualSendAllEnd {
		s.sendAllEnd = true
		s.manualSendAllEnd = false
		s.wakeup()
	}
}

func (s *SendRecv) isRemoteWanted(name string) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	_, found := s.remoteWanted[name]
	return found
}

func (s *SendRecv) send(stream StreamNotify) {
	var err error
	var notice edgeproto.Notice

	sendAll := true
	sendAllSpan := log.StartSpan(log.DebugLevelNotify, "notify-send-all")
	sendAllSpan.SetTag("peerAddr", s.peerAddr)
	sendAllSpan.SetTag("peer", s.peer)
	sendAllSpan.SetTag("cliserv", s.cliserv)
	sendAllCtx := opentracing.ContextWithSpan(context.Background(), sendAllSpan)

	// trigger initial sendAll
	s.wakeup()
	streamDone := false

	for !s.done && err == nil {
		// Select with channels is used here rather than a condition
		// variable to be able to detect when the underlying connection
		// is done/cancelled, as the only way to detect that is via a
		// channel, and you can't mix waiting on condition variables
		// and channels.
		select {
		case <-s.signal:
		case <-stream.Context().Done():
			err = stream.Context().Err()
			streamDone = true
		}
		if streamDone || s.done {
			break
		}
		hasData := false
		// UpdateAll/PrepData in reverse order, otherwise
		// we may capture a dependent object update without
		// capturing the earlier-in-the-loop object that
		// it depended on.
		for ii := len(s.sendlist) - 1; ii >= 0; ii-- {
			if sendAll {
				s.sendlist[ii].UpdateAll(sendAllCtx)
			}
			if s.sendlist[ii].PrepData() {
				hasData = true
			}
		}
		if !hasData && !s.done && !sendAll && !s.sendAllEnd {
			continue
		}
		if s.done {
			break
		}
		if sendAll {
			log.SpanLog(sendAllCtx, log.DebugLevelNotify, "send all", "peer", s.peer)
			s.stats.SendAll++
		}
		// Note that order is important here, as some objects
		// may have dependencies on other objects. It's up to the
		// caller to make sure CacheSend objects are registered
		// in the desired send order.
		for _, send := range s.sendlist {
			err = send.Send(stream, &notice, s.peerAddr)
			if err != nil {
				break
			}
		}
		if err != nil {
			break
		}
		if s.sendAllEnd {
			s.sendAllEnd = false
			log.SpanLog(sendAllCtx, log.DebugLevelNotify, "send all end", "peer", s.peer)
			notice.Action = edgeproto.NoticeAction_SENDALL_END
			notice.Any = types.Any{}
			err = stream.Send(&notice)
			if err != nil {
				log.SpanLog(sendAllCtx, log.DebugLevelNotify,
					"send all end", "peer", s.peer, "err", err)
				break
			}
			sendAllSpan.Finish()
			sendAllSpan = nil
		}
		if sendAll {
			sendAll = false
		}
		if err != nil {
			break
		}
	}
	if sendAllSpan != nil {
		sendAllSpan.Finish()
	}
	close(s.sendRunning)
}

func (s *SendRecv) recv(stream StreamNotify, notifyId int64, cleanup Cleanup) {
	recvAll := true
	// Note we never actually receive a SendAll_Start message,
	// it's implicit on the rx side when the connection is established.
	sendAllRecv := s.sendAllRecvHandler
	if sendAllRecv != nil {
		sendAllRecv.RecvAllStart()
	}

	for _, recv := range s.recvmap {
		recv.RecvAllStart()
	}
	for !s.done {
		notice, err := stream.Recv()
		if s.done {
			break
		}
		if err != nil {
			log.DebugLog(log.DebugLevelNotify,
				fmt.Sprintf("%s receive", s.cliserv), "err", err)
			break
		}
		func() {
			// anonymous inner func so we can use defer to close span
			ctx := context.Background()
			name, err := types.AnyMessageName(&notice.Any)
			if notice.Span != "" {
				spanName := fmt.Sprintf("notify-recv %s", name)
				span := log.NewSpanFromString(log.DebugLevelNotify, notice.Span, spanName)
				span.SetTag("action", notice.Action)
				span.SetTag("cliserv", s.cliserv)
				span.SetTag("peer", s.peerAddr)
				defer span.Finish()
				ctx = opentracing.ContextWithSpan(ctx, span)
			}
			if err != nil && notice.Action != edgeproto.NoticeAction_SENDALL_END {
				log.SpanLog(ctx, log.DebugLevelNotify, "hit error", "peer", s.peer, "err", err)
				return
			}
			if recvAll && notice.Action == edgeproto.NoticeAction_SENDALL_END {
				for _, recv := range s.recvmap {
					recv.RecvAllEnd(ctx, cleanup)
				}
				sendAllRecv := s.sendAllRecvHandler
				if sendAllRecv != nil {
					sendAllRecv.RecvAllEnd(ctx)
				}
				recvAll = false
				return
			}
			recv := s.recvmap[name]
			if recv != nil {
				recv.Recv(ctx, notice, notifyId, s.peerAddr)
			} else {
				log.DebugLog(log.DebugLevelNotify,
					fmt.Sprintf("%s recv unhandled", s.cliserv),
					"peerAddr", s.peerAddr,
					"peer", s.peer,
					"action", notice.Action,
					"name", name)
			}
		}()
	}
	close(s.recvRunning)
}

func (s *SendRecv) wakeup() {
	// This puts true in the channel unless it is full,
	// then the default (noop) case is performed.
	// The signal channel is used to tell the thread to run.
	// It is a replacement for a condition variable, which
	// we cannot use (see comments in Server send())
	select {
	case s.signal <- true:
	default:
	}
}

func (s *SendRecv) hasCloudletKey(key *edgeproto.CloudletKey) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	_, found := s.cloudletKeys[*key]
	return found
}

func (s *SendRecv) updateCloudletKey(action edgeproto.NoticeAction, key *edgeproto.CloudletKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if action == edgeproto.NoticeAction_UPDATE {
		log.DebugLog(log.DebugLevelNotify, "sendrecv add cloudletkey", "key", key)
		s.cloudletKeys[*key] = struct{}{}
	} else {
		log.DebugLog(log.DebugLevelNotify, "sendrecv remove cloudletkey", "key", key)
		delete(s.cloudletKeys, *key)
	}
}

func (s *SendRecv) setObjStats(stats *Stats) {
	stats.ObjSend = make(map[string]uint64)
	stats.ObjRecv = make(map[string]uint64)
	for _, send := range s.sendlist {
		stats.ObjSend[send.GetName()] = send.GetSendCount()
	}
	for _, recv := range s.recvmap {
		stats.ObjRecv[recv.GetName()] = recv.GetRecvCount()
	}
}
