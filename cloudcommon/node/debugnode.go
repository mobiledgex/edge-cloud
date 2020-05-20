package node

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

var DefaultDebugTimeout = 10 * time.Second

type DebugFunc func(ctx context.Context, req *edgeproto.DebugRequest) string

const (
	EnableDebugLevels    = "enable-debug-levels"
	DisableDebugLevels   = "disable-debug-levels"
	ShowDebugLevels      = "show-debug-levels"
	RefreshInternalCerts = "refresh-internal-certs"
	StartCpuProfileCmd   = "start-cpu-profile"
	StopCpuProfileCmd    = "stop-cpu-profile"
	GetMemProfileCmd     = "get-mem-profile"
)

type DebugNode struct {
	mgr         *NodeMgr
	sendReply   *notify.DebugReplySend
	sendRequest *notify.DebugRequestSendMany
	funcs       map[string]DebugFunc
	requests    map[uint64]*debugCall
	mux         sync.Mutex
}

type debugCall struct {
	id        uint64
	sendNodes map[edgeproto.NodeKey]struct{}
	done      chan bool
	callReply func(ctx context.Context, reply *edgeproto.DebugReply)
}

func (s *DebugNode) Init(mgr *NodeMgr) {
	s.mgr = mgr
	s.requests = make(map[uint64]*debugCall)
	s.funcs = make(map[string]DebugFunc)
	s.AddDebugFunc(EnableDebugLevels, enableDebugLevels)
	s.AddDebugFunc(DisableDebugLevels, disableDebugLevels)
	s.AddDebugFunc(ShowDebugLevels, showDebugLevels)
	s.AddDebugFunc(RefreshInternalCerts,
		func(ctx context.Context, req *edgeproto.DebugRequest) string {
			mgr.InternalPki.triggerRefresh()
			return "triggered refresh"
		})
	s.AddDebugFunc(StartCpuProfileCmd,
		func(ctx context.Context, req *edgeproto.DebugRequest) string {
			return StartCpuProfile()
		})
	s.AddDebugFunc(StopCpuProfileCmd,
		func(ctx context.Context, req *edgeproto.DebugRequest) string {
			return StopCpuProfile()
		})
	s.AddDebugFunc(GetMemProfileCmd,
		func(ctx context.Context, req *edgeproto.DebugRequest) string {
			return GetMemProfile()
		})
}

func (s *DebugNode) AddDebugFunc(cmd string, f DebugFunc) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.funcs[cmd] = f
}

func (s *DebugNode) RegisterClient(client *notify.Client) {
	s.sendReply = notify.NewDebugReplySend()
	client.RegisterRecv(notify.NewDebugRequestRecv(s))
	client.RegisterSend(s.sendReply)
}

func (s *DebugNode) RegisterServer(server *notify.ServerMgr) {
	s.sendRequest = notify.NewDebugRequestSendMany()
	server.RegisterSend(s.sendRequest)
	server.RegisterRecv(notify.NewDebugReplyRecvMany(s))
}

// Handle DebugRequest received via the notify framework.
// Replies are sent back up the notify connection.
func (s *DebugNode) RecvDebugRequest(ctx context.Context, req *edgeproto.DebugRequest) {
	go s.handleRequest(ctx, req, func(ctx context.Context, reply *edgeproto.DebugReply) {
		s.sendReply.Update(ctx, reply)
	})
}

// Handle DebugRequest via grpc API call.
// Replies are sent back to the grpc client.
func (s *DebugNode) DebugRequest(req *edgeproto.DebugRequest, cb edgeproto.DebugApi_RunDebugServer) error {
	req.Id = rand.Uint64()
	if req.Timeout == 0 {
		req.Timeout = edgeproto.Duration(DefaultDebugTimeout)
	}
	return s.handleRequest(cb.Context(), req, func(ctx context.Context, reply *edgeproto.DebugReply) {
		reply.Id = 0
		cb.Send(reply)
	})
}

func (s *DebugNode) handleRequest(ctx context.Context, req *edgeproto.DebugRequest, callReply func(ctx context.Context, reply *edgeproto.DebugReply)) error {
	// run local first
	if nodeMatches(&s.mgr.MyNode.Key, &req.Node) {
		reply := edgeproto.DebugReply{}
		reply.Node = s.mgr.MyNode.Key
		reply.Id = req.Id
		reply.Output = s.RunDebug(ctx, req)
		callReply(ctx, &reply)
	}

	if s.sendRequest == nil {
		// no possible children
		return nil
	}

	call := debugCall{}
	call.id = req.Id
	call.done = make(chan bool)
	call.sendNodes = make(map[edgeproto.NodeKey]struct{})
	call.callReply = callReply
	// Only send to children with matching node keys
	// The local NodeCache has every node below this node in it.
	// However, the notifyId with each entry is for the immediate child
	// only, even if the node is multiple levels down the tree.
	// So we can figure out which children to forwards requests to based
	// on the notifyId, and we can track every response we expect based
	// on the Node Key match from the NodeCache entries.
	sendNotifyIds := make(map[int64]struct{})
	cache := &s.mgr.NodeCache
	cache.Mux.Lock()
	for key, data := range cache.Objs {
		node := data.Obj
		if nodeMatches(&key, &req.Node) {
			if node.NotifyId == 0 {
				// this is me, already handled above
				continue
			}
			sendNotifyIds[node.NotifyId] = struct{}{}
			call.sendNodes[key] = struct{}{}
		}
	}
	cache.Mux.Unlock()
	numNodes := len(call.sendNodes)

	s.mux.Lock()
	s.requests[req.Id] = &call
	s.mux.Unlock()

	count := s.sendRequest.UpdateFiltered(ctx, req, func(ctx context.Context, send *notify.DebugRequestSend, msg *edgeproto.DebugRequest) bool {
		_, found := sendNotifyIds[send.GetNotifyId()]
		return found
	})
	log.SpanLog(ctx, log.DebugLevelApi, "handleRequest sent to children", "id", req.Id, "count", count, "numNodes", numNodes)
	timedout := false
	if count > 0 {
		select {
		case <-call.done:
		case <-time.After(req.Timeout.TimeDuration()):
			timedout = true
		}
	}

	s.mux.Lock()
	delete(s.requests, req.Id)
	s.mux.Unlock()

	if timedout {
		for nodeKey, _ := range call.sendNodes {
			reply := edgeproto.DebugReply{}
			reply.Node = nodeKey
			reply.Id = req.Id
			reply.Output = "request timed out"
			callReply(ctx, &reply)
		}
		log.SpanLog(ctx, log.DebugLevelApi, "handleRequest timed out", "remaining", count, "remaining-nodes", call.sendNodes)
	}
	return nil
}

// Handle replies from notify children
func (s *DebugNode) RecvDebugReply(ctx context.Context, reply *edgeproto.DebugReply) {
	s.mux.Lock()
	call, found := s.requests[reply.Id]
	s.mux.Unlock()
	if !found {
		log.SpanLog(ctx, log.DebugLevelApi, "unregistered DebugReply recv", "reply", reply)
		return
	}

	done := false
	s.mux.Lock()
	delete(call.sendNodes, reply.Node)
	if len(call.sendNodes) == 0 {
		done = true
	}
	log.SpanLog(ctx, log.DebugLevelApi, "recv reply", "id", reply.Id, "numNodes", len(call.sendNodes), "node", reply.Node)
	s.mux.Unlock()

	if s.mgr.NodeCache.setRegion != "" {
		reply.Node.Region = s.mgr.NodeCache.setRegion
	}

	call.callReply(ctx, reply)
	if done {
		close(call.done)
	}
}

func (s *DebugNode) RunDebug(ctx context.Context, req *edgeproto.DebugRequest) string {
	s.mux.Lock()
	f, ok := s.funcs[req.Cmd]
	if !ok {
		cmdstrs := make([]string, 0)
		for cmd, _ := range s.funcs {
			cmdstrs = append(cmdstrs, cmd)
		}
		s.mux.Unlock()
		sort.Strings(cmdstrs)
		return fmt.Sprintf("Unknown cmd %s, cmds are %s", req.Cmd, strings.Join(cmdstrs, ","))
	}
	s.mux.Unlock()
	return f(ctx, req)
}

func enableDebugLevels(ctx context.Context, req *edgeproto.DebugRequest) string {
	log.SetDebugLevelStrs(req.Levels)
	return "enabled debug levels, now " + log.GetDebugLevelStrs()
}

func disableDebugLevels(ctx context.Context, req *edgeproto.DebugRequest) string {
	log.ClearDebugLevelStrs(req.Levels)
	return "disabled debug levels, now " + log.GetDebugLevelStrs()
}

func showDebugLevels(ctx context.Context, req *edgeproto.DebugRequest) string {
	return log.GetDebugLevelStrs()
}
