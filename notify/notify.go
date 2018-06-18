// Update DME (distributed matching engines) with changes in app inst data

// Why do this instead of having receivers just do etcd watch?
// 1. Receivers load the controller instances instead of etcd instances (many
// controller instances vs 3 or max 7 etcd instances).
// 2. We can control load balancing - sharding which receivers are serviced by
// which controllers. With etcd watch we don't have control over which etcd
// instance each receiver connects to. This is TODO.
// 3. We can send more than just etcd database objects.
// 4. Based on github comments from 11/2017, the etcd watch reconnect behavior
// is not well defined and users have been having trouble with it.
// 5. Watch is a pull protocol. The client connects to etcd to receive updates.
// We really want a push protocol, where the client is sending the updates.
// That way if a send fails, the client can reconnect. In the watch case,
// the server is sending the updates, and on send failure, has no way to tell
// the client that it needs to reconnect to the server.

package notify

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

var NotifyRetryTime time.Duration = 250 * time.Millisecond

const NotifyVersion uint32 = 1

type ServerHandler interface {
	// Get all the keys for known app insts.
	// The value associated with the key is ignored.
	GetAllAppInstKeys(keys map[edgeproto.AppInstKey]struct{})
	// Copy back the value for the app inst.
	// If the app inst was not found, return false instead of true.
	GetAppInst(key *edgeproto.AppInstKey, buf *edgeproto.AppInst) bool
	// Get all the keys for known cloudlets.
	// The value associated with the key is ignored.
	GetAllCloudletKeys(keys map[edgeproto.CloudletKey]struct{})
	// Copy back the value for the cloudlet.
	// If the cloudlet was not found, return false instead of true.
	GetCloudlet(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool
}

type ServerStats struct {
	AppInstsSent  uint64
	CloudletsSent uint64
}

// Server is on the controller side and sends data to DME/CRM clients.
// On first connect, it will send all data from the database that is
// required by the client. After that it will send objects only when
// they are changed.
type Server struct {
	peerAddr  string
	appInsts  map[edgeproto.AppInstKey]struct{}
	cloudlets map[edgeproto.CloudletKey]struct{}
	mux       util.Mutex
	signal    chan bool
	done      bool
	handler   ServerHandler
	stats     ServerStats
	version   uint32
	requestor edgeproto.NoticeRequestor
	running   chan struct{}
}

type ServerMgr struct {
	table   map[string]*Server
	mux     util.Mutex
	handler ServerHandler
	serv    *grpc.Server
}

var serverMgr ServerMgr

// Keepalive parameters to close the connection if the other end
// goes away unexpectedly. The server and client parameters must be balanced
// correctly or the connection may be closed incorrectly.
const (
	infinity   = time.Duration(math.MaxInt64)
	kpInterval = 30 * time.Second
)

var serverParams = keepalive.ServerParameters{
	MaxConnectionIdle:     3 * kpInterval,
	MaxConnectionAge:      infinity,
	MaxConnectionAgeGrace: infinity,
	Time:    kpInterval,
	Timeout: kpInterval,
}
var clientParams = keepalive.ClientParameters{
	Time:    kpInterval,
	Timeout: kpInterval,
}
var serverEnforcement = keepalive.EnforcementPolicy{
	MinTime: 1 * time.Second,
}

func ServerMgrStart(addr string, handler ServerHandler) {
	serverMgr.mux.Lock()
	defer serverMgr.mux.Unlock()

	if serverMgr.table != nil {
		return
	}
	serverMgr.table = make(map[string]*Server)
	serverMgr.handler = handler

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		util.FatalLog("ServerMgr listen failed", "err", err)
	}
	serverMgr.serv = grpc.NewServer(grpc.KeepaliveParams(serverParams), grpc.KeepaliveEnforcementPolicy(serverEnforcement))
	edgeproto.RegisterNotifyApiServer(serverMgr.serv, &serverMgr)
	util.DebugLog(util.DebugLevelNotify, "ServerMgr listening", "addr", addr)
	go func() {
		err = serverMgr.serv.Serve(lis)
		if err != nil {
			util.FatalLog("ServerMgr serve failed", "err", err)
		}
	}()
}

func ServerMgrDone() {
	serverMgr.mux.Lock()
	serverMgr.serv.Stop()
	if serverMgr.table != nil {
		for _, server := range serverMgr.table {
			server.Stop()
		}
	}
	serverMgr.table = nil
	serverMgr.handler = nil
	serverMgr.mux.Unlock()
}

func (s *ServerMgr) StreamNotice(stream edgeproto.NotifyApi_StreamNoticeServer) error {
	ctx := stream.Context()
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return errors.New("Notify ServerMgr unable to get peer context")
	}
	peerAddr := peer.Addr.String()

	server := Server{}
	server.peerAddr = peerAddr
	server.appInsts = make(map[edgeproto.AppInstKey]struct{})
	server.cloudlets = make(map[edgeproto.CloudletKey]struct{})
	server.signal = make(chan bool, 1)
	server.handler = serverMgr.handler
	server.running = make(chan struct{})
	// wakeup makes sure server does send all
	server.wakeup()

	s.mux.Lock()
	s.table[peerAddr] = &server
	s.mux.Unlock()

	err := server.StreamNotice(stream)

	s.mux.Lock()
	// another connect may come in from the same client so do not
	// remove it unless it's the same one.
	if remove, _ := s.table[peerAddr]; remove == &server {
		delete(s.table, peerAddr)
	}
	s.mux.Unlock()
	return err
}

func GetServerStats(peerAddr string) *ServerStats {
	stats := &ServerStats{}
	serverMgr.mux.Lock()
	defer serverMgr.mux.Unlock()
	if peerAddr != "" {
		server, found := serverMgr.table[peerAddr]
		if found {
			*stats = server.stats
		}
	}
	return stats
}

func UpdateAppInst(key *edgeproto.AppInstKey) {
	serverMgr.mux.Lock()
	defer serverMgr.mux.Unlock()
	for _, server := range serverMgr.table {
		server.UpdateAppInst(key)
	}
}

func UpdateCloudlet(key *edgeproto.CloudletKey) {
	serverMgr.mux.Lock()
	defer serverMgr.mux.Unlock()
	for _, server := range serverMgr.table {
		if server.requestor != edgeproto.NoticeRequestor_NoticeRequestorCRM {
			continue
		}
		server.UpdateCloudlet(key)
	}
}

func (s *Server) wakeup() {
	// This puts true in the channel unless it is full,
	// then the default (noop) case is performed.
	// The signal channel is used to tell the thread to run.
	// It is a replacement for a condition variable, which
	// we cannot use (see comments in Server Run())
	select {
	case s.signal <- true:
	default:
	}
}

func (s *Server) UpdateAppInst(key *edgeproto.AppInstKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.appInsts[*key] = struct{}{}
	s.wakeup()
}

func (s *Server) UpdateCloudlet(key *edgeproto.CloudletKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.cloudlets[*key] = struct{}{}
	s.wakeup()
}

func (s *Server) StreamNotice(stream edgeproto.NotifyApi_StreamNoticeServer) error {
	var req *edgeproto.NoticeRequest
	var err error
	var notice edgeproto.NoticeReply
	var noticeAppInst edgeproto.NoticeReply_AppInst
	var noticeCloudlet edgeproto.NoticeReply_Cloudlet
	var appInst edgeproto.AppInst
	var cloudlet edgeproto.Cloudlet

	noticeAppInst.AppInst = &appInst
	noticeCloudlet.Cloudlet = &cloudlet
	defer func() {
		// handle failure
		if err != nil {
			st, ok := status.FromError(err)
			if err == context.Canceled || (ok && st.Code() == codes.Canceled) {
				util.DebugLog(util.DebugLevelNotify, "Notify server connection closed", "client", s.peerAddr, "err", err)
			} else {
				util.InfoLog("Notify server connection failed", "client", s.peerAddr, "err", err)
			}
		}

	}()

	// initial connection is version exchange
	// this also sets the connection Id so we can ignore spurious old
	// buffered messages
	req, err = stream.Recv()
	if err != nil {
		return err
	}
	if req.Action != edgeproto.NoticeAction_VERSION {
		util.DebugLog(util.DebugLevelNotify, "Notify server bad action", "expected", edgeproto.NoticeAction_VERSION, "got", notice.Action)
		return errors.New("Notify server expected action version")
	}
	if req.Requestor != edgeproto.NoticeRequestor_NoticeRequestorDME && req.Requestor != edgeproto.NoticeRequestor_NoticeRequestorCRM {
		return errors.New("Notify server bad requestor value")
	}
	// set requestor type
	s.requestor = req.Requestor
	// use lowest common version
	if req.Version > NotifyVersion {
		s.version = req.Version
	} else {
		s.version = NotifyVersion
	}
	// send back my version
	notice.Action = edgeproto.NoticeAction_VERSION
	notice.Version = s.version
	err = stream.Send(&notice)
	if err != nil {
		return err
	}
	util.DebugLog(util.DebugLevelNotify, "Notify server connected", "client", s.peerAddr, "version", s.version, "supported-version", NotifyVersion)

	sendAll := true
	for !s.done {
		// Select with channels is used here rather than a condition
		// variable to be able to detect when the underlying connection
		// is done/cancelled, as the only way to detect that is via a
		// channel, and you can't mix waiting on condition variables
		// and channels.
		select {
		case <-s.signal:
		case <-stream.Context().Done():
			err = stream.Context().Err()
			s.done = true
		}
		s.mux.Lock()
		if len(s.appInsts) == 0 && len(s.cloudlets) == 0 && !s.done && !sendAll && stream.Context().Err() == nil {
			s.mux.Unlock()
			continue
		}
		appInsts := s.appInsts
		cloudlets := s.cloudlets
		s.appInsts = make(map[edgeproto.AppInstKey]struct{})
		s.cloudlets = make(map[edgeproto.CloudletKey]struct{})
		s.mux.Unlock()
		if s.done {
			break
		}
		if sendAll {
			util.DebugLog(util.DebugLevelNotify, "Send all", "client", s.peerAddr)
			s.handler.GetAllAppInstKeys(appInsts)
			if s.requestor == edgeproto.NoticeRequestor_NoticeRequestorCRM {
				s.handler.GetAllCloudletKeys(cloudlets)
			}
		}
		// send appInsts
		notice.Data = &noticeAppInst
		for key, _ := range appInsts {
			found := s.handler.GetAppInst(&key, &appInst)
			if found {
				notice.Action = edgeproto.NoticeAction_UPDATE
			} else {
				notice.Action = edgeproto.NoticeAction_DELETE
				appInst.Key = key
			}
			util.DebugLog(util.DebugLevelNotify, "Send app inst", "client", s.peerAddr, "action", notice.Action, "key", appInst.Key.GetKeyString())
			err = stream.Send(&notice)
			if err != nil {
				break
			}
			s.stats.AppInstsSent++
		}
		// send cloudlets
		notice.Data = &noticeCloudlet
		for key, _ := range cloudlets {
			found := s.handler.GetCloudlet(&key, &cloudlet)
			if found {
				notice.Action = edgeproto.NoticeAction_UPDATE
			} else {
				notice.Action = edgeproto.NoticeAction_DELETE
				cloudlet.Key = key
			}
			util.DebugLog(util.DebugLevelNotify, "Send cloudlet", "client", s.peerAddr, "action", notice.Action, "key", cloudlet.Key.GetKeyString())
			err = stream.Send(&notice)
			if err != nil {
				break
			}
			s.stats.CloudletsSent++
		}
		if sendAll && err == nil {
			notice.Action = edgeproto.NoticeAction_SENDALL_END
			notice.Data = nil
			err = stream.Send(&notice)
			if err != nil {
				break
			}
			sendAll = false
		}
		appInsts = nil
		cloudlets = nil
	}
	close(s.running)
	return err
}

func (s *Server) Stop() {
	s.mux.Lock()
	s.done = true
	s.wakeup()
	s.mux.Unlock()
	<-s.running
}

type AllMaps struct {
	AppInsts  map[edgeproto.AppInstKey]struct{}
	Cloudlets map[edgeproto.CloudletKey]struct{}
}

type ClientHandler interface {
	// NotifySendAllMaps contains all the keys that were updated in
	// the initial send of all data. Any keys that were not sent should
	// be purged from local memory since they no longer exist on the
	// sender.
	HandleSendAllDone(allMaps *AllMaps)
	// Handle an update or delete notice from the sender
	HandleNotice(reply *edgeproto.NoticeReply) error
}

type ClientStats struct {
	Connects uint64
}

type Client struct {
	addrs     []string
	handler   ClientHandler
	requestor edgeproto.NoticeRequestor
	version   uint32
	stats     ClientStats
	mux       util.Mutex
	done      bool
	running   chan struct{}
	cancel    context.CancelFunc
	localAddr string
}

func cancelNoop() {}

func NewDMEClient(addrs []string, handler ClientHandler) *Client {
	return newClient(addrs, handler, edgeproto.NoticeRequestor_NoticeRequestorDME)
}

func NewCRMClient(addrs []string, handler ClientHandler) *Client {
	return newClient(addrs, handler, edgeproto.NoticeRequestor_NoticeRequestorCRM)
}

func newClient(addrs []string, handler ClientHandler, requestor edgeproto.NoticeRequestor) *Client {
	s := &Client{}
	s.addrs = addrs
	s.handler = handler
	s.requestor = requestor
	return s
}

func (s *Client) Run() {
	var conn *grpc.ClientConn
	var request edgeproto.NoticeRequest
	var reply *edgeproto.NoticeReply
	var err error
	var commErr error
	var stream edgeproto.NotifyApi_StreamNoticeClient

	s.done = false
	s.cancel = cancelNoop
	s.SetLocalAddr("")
	s.running = make(chan struct{})
	addrIdx := rand.Int() % len(s.addrs)
	tries := 0

	cleanup := func() {
		if conn != nil {
			err = conn.Close()
		}
		s.cancel()
		s.cancel = cancelNoop
		s.SetLocalAddr("")
		stream = nil
		commErr = nil
	}

	for !s.done {
		if commErr != nil {
			util.DebugLog(util.DebugLevelNotify, "Notify client communication err", "addr", s.addrs[addrIdx], "local", s.GetLocalAddr(), "error", commErr)
			cleanup()
		}
		if stream == nil {
			// connect to server
			tries++
			addrIdx++
			if addrIdx == len(s.addrs) {
				addrIdx = 0
			}
			ctx, cancel := context.WithTimeout(context.Background(), NotifyRetryTime)
			conn, err = grpc.DialContext(ctx, s.addrs[addrIdx], grpc.WithInsecure(), grpc.WithStatsHandler(&grpcStatsHandler{client: s}), grpc.WithKeepaliveParams(clientParams))
			cancel()
			if err != nil {
				if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
					time.Sleep(NotifyRetryTime)
				}
				continue
			}
			api := edgeproto.NewNotifyApiClient(conn)
			ctx, cancel = context.WithCancel(context.Background())
			stream, err = api.StreamNotice(ctx)
			s.cancel = cancel
			if err != nil {
				util.DebugLog(util.DebugLevelNotify, "Notify client get client", "addr", s.addrs[addrIdx], "error", err)
				cleanup()
				time.Sleep(NotifyRetryTime)
				continue
			}
			s.stats.Connects++

			// Send our version and read back remote version.
			// We use the lowest common version.
			request.Version = NotifyVersion
			request.Action = edgeproto.NoticeAction_VERSION
			request.Requestor = s.requestor
			commErr = stream.Send(&request)
			if commErr != nil {
				continue
			}
			reply, commErr = stream.Recv()
			if commErr != nil {
				continue
			}
			if request.Version > reply.Version {
				s.version = reply.Version
			} else {
				s.version = request.Version
			}
			util.DebugLog(util.DebugLevelNotify, "Notify client connected", "addr", s.addrs[addrIdx], "local", s.GetLocalAddr(), "version", s.version, "supported-version", NotifyVersion, "tries", tries)
			tries = 0
		}
		// server will send all data first
		sendAllMaps := &AllMaps{}
		sendAllMaps.AppInsts = make(map[edgeproto.AppInstKey]struct{})
		sendAllMaps.Cloudlets = make(map[edgeproto.CloudletKey]struct{})
		for !s.done {
			reply, commErr = stream.Recv()
			if s.done {
				break
			}
			if commErr != nil {
				break
			}
			if sendAllMaps != nil && reply.Action == edgeproto.NoticeAction_UPDATE {
				appInst := reply.GetAppInst()
				if appInst != nil {
					sendAllMaps.AppInsts[appInst.Key] = struct{}{}
				}
				cloudlet := reply.GetCloudlet()
				if cloudlet != nil {
					sendAllMaps.Cloudlets[cloudlet.Key] = struct{}{}
				}
			}
			if reply.Action == edgeproto.NoticeAction_SENDALL_END {
				s.handler.HandleSendAllDone(sendAllMaps)
				sendAllMaps = nil
				continue
			}
			if reply.Action != edgeproto.NoticeAction_UPDATE && reply.Action != edgeproto.NoticeAction_DELETE {
				commErr = errors.New("Unexpected notice action, not update or delete")
				break
			}
			commErr = s.handler.HandleNotice(reply)
			if commErr != nil {
				break
			}
		}
	}
	util.DebugLog(util.DebugLevelNotify, "Notify client cancelled", "local", s.GetLocalAddr())
	if stream != nil {
		cleanup()
	}
	close(s.running)
}

func (s *Client) Stop() {
	s.mux.Lock()
	s.done = true
	s.cancel()
	s.mux.Unlock()
	if s.running != nil {
		<-s.running
	}
}

func (s *Client) GetStats() *ClientStats {
	stats := &ClientStats{}
	*stats = s.stats
	return stats
}

func (s *Client) SetLocalAddr(addr string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.localAddr = addr
}

func (s *Client) GetLocalAddr() string {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.localAddr
}

type grpcStatsHandler struct {
	client *Client
}

func (s *grpcStatsHandler) TagRPC(ctx context.Context, tag *stats.RPCTagInfo) context.Context {
	return ctx
}

func (s *grpcStatsHandler) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {}

func (s *grpcStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	s.client.SetLocalAddr(info.LocalAddr.String())
	return ctx
}

func (s *grpcStatsHandler) HandleConn(ctx context.Context, connStats stats.ConnStats) {}
