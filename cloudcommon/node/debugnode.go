package node

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	yaml "github.com/mobiledgex/yaml/v2"
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
	DisableSampleLog     = "disable-sample-logging"
	EnableSampleLog      = "enable-sample-logging"
	DumpCloudletPools    = "dump-cloudlet-pools"
	DumpStackTrace       = "dump-stack-trace"
	DumpNotifyConns      = "dump-notify-state"
)

type DebugNode struct {
	mgr              *NodeMgr
	sendReply        *notify.DebugReplySend
	sendRequest      *notify.DebugRequestSendMany
	funcs            map[string]DebugFunc
	requests         map[uint64]*debugCall
	mux              sync.Mutex
	notifyClients    []*notify.Client
	notifyServerMgrs []*notify.ServerMgr
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
	s.notifyClients = make([]*notify.Client, 0)
	s.notifyServerMgrs = make([]*notify.ServerMgr, 0)
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
	s.AddDebugFunc(DisableSampleLog, disableSampledLogging)
	s.AddDebugFunc(EnableSampleLog, enableSampledLogging)
	s.AddDebugFunc(DumpCloudletPools,
		func(ctx context.Context, req *edgeproto.DebugRequest) string {
			dat := mgr.CloudletPoolLookup.Dumpable()
			out, err := json.Marshal(dat)
			if err != nil {
				return err.Error()
			}
			return string(out)
		})
	s.AddDebugFunc(DumpStackTrace, dumpStackTrace)
	s.AddDebugFunc(DumpNotifyConns, s.dumpNotifyState)
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
	s.mux.Lock()
	s.notifyClients = append(s.notifyClients, client)
	s.mux.Unlock()
}

func (s *DebugNode) RegisterServer(server *notify.ServerMgr) {
	s.sendRequest = notify.NewDebugRequestSendMany()
	server.RegisterSend(s.sendRequest)
	server.RegisterRecv(notify.NewDebugReplyRecvMany(s))
	s.mux.Lock()
	s.notifyServerMgrs = append(s.notifyServerMgrs, server)
	s.mux.Unlock()
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
	if req.Cmd == "" {
		return fmt.Errorf("no cmd specified")
	}
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

func disableSampledLogging(ctx context.Context, req *edgeproto.DebugRequest) string {
	log.SamplingEnabled = false
	return "disabled log sampling"
}

func enableSampledLogging(ctx context.Context, req *edgeproto.DebugRequest) string {
	log.SamplingEnabled = true
	return "enabled log sampling"
}

func dumpStackTrace(ctx context.Context, req *edgeproto.DebugRequest) string {
	help := "Args should indicate buffer size in KB and whether to write to log file, i.e. 100,false. Defaults are 100,false."
	if strings.ToLower(req.Args) == "help" {
		return help
	}
	// Default to 100Kb stack trace buffer.
	// Controller may have lots of threads, QA controller
	// had a stack trace that used up 513Kb.
	bufSizeKb := 100
	printToLog := false
	args := strings.Split(req.Args, ",")
	if len(args) > 0 {
		if val, err := strconv.Atoi(args[0]); err == nil {
			bufSizeKb = val
		} else {
			return help
		}
	}
	if len(args) > 1 {
		if b, err := strconv.ParseBool(args[1]); err == nil {
			printToLog = b
		} else {
			return help
		}
	}
	buf := make([]byte, bufSizeKb*1024)
	runtime.Stack(buf, true)
	if printToLog {
		fmt.Println(string(buf))
	}
	return string(buf)
}

type NotifyState struct {
	ClientStates    []*notify.ClientState
	ServerMgrStates []*notify.ServerMgrState
}

func (s *DebugNode) dumpNotifyState(ctx context.Context, req *edgeproto.DebugRequest) string {
	s.mux.Lock()
	state := NotifyState{}
	for _, client := range s.notifyClients {
		state.ClientStates = append(state.ClientStates, client.GetState())
	}
	for _, serverMgr := range s.notifyServerMgrs {
		state.ServerMgrStates = append(state.ServerMgrStates, serverMgr.GetState())
	}
	s.mux.Unlock()
	out, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Sprintf("Failed to marshal data, %s", err)
	}
	return string(out)
}
