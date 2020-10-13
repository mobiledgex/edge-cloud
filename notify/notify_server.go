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
	"crypto/tls"
	"errors"
	"net"
	"reflect"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var NotifyRetryTime time.Duration = 250 * time.Millisecond

const NotifyVersion uint32 = 1

// Server is on the upstream side and sends data to downstream clients.
// On first connect, it will send all data from the database that is
// required by the client. After that it will send objects only when
// they are changed.
type Server struct {
	sendrecv SendRecv
	peerAddr string
	mux      util.Mutex
	version  uint32
	notifyId int64
	running  chan struct{}
}

// ServerMgr maintains all the Server threads for clients connected to us.
type ServerMgr struct {
	table    map[string]*Server
	mux      util.Mutex
	sends    []NotifySendMany
	recvs    []NotifyRecvMany
	notifyId int64
	serv     *grpc.Server
	name     string
	regServ  func(s *grpc.Server)
}

// NotifySendMany and NotifyRecvMany are implemented by auto-generated code.
// They are simply thin layers which can create new NotifySend/NotifyRecv
// implementations which are object-specific. These are required by the
// ServerMgr which needs to generate NotifySend/NotifyRecv objects for each
// new connection from a new client.

type NotifySendMany interface {
	// Allocate a new Send object
	NewSend(peerAddr string, notifyId int64) NotifySend
	// Free a Send object
	DoneSend(peerAddr string, send NotifySend)
}

type NotifyRecvMany interface {
	// Allocate a new Recv object
	NewRecv() NotifyRecv
	// Flush stale data for the connection after disconnect
	Flush(ctx context.Context, notifyId int64)
}

var ServerMgrOne ServerMgr

func (mgr *ServerMgr) RegisterSend(sendMany NotifySendMany) {
	mgr.mux.Lock()
	defer mgr.mux.Unlock()
	mgr.sends = append(mgr.sends, sendMany)
}

func (mgr *ServerMgr) RegisterRecv(recvMany NotifyRecvMany) {
	mgr.mux.Lock()
	defer mgr.mux.Unlock()
	mgr.recvs = append(mgr.recvs, recvMany)
}

func (mgr *ServerMgr) RegisterServerCb(registerServer func(s *grpc.Server)) {
	mgr.regServ = registerServer
}

func (mgr *ServerMgr) Start(name, addr string, tlsConfig *tls.Config) {
	mgr.mux.Lock()
	defer mgr.mux.Unlock()

	if mgr.table != nil {
		return
	}
	mgr.name = name
	mgr.notifyId = 1
	mgr.table = make(map[string]*Server)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.FatalLog("ServerMgr listen failed", "err", err)
	}

	mgr.serv = grpc.NewServer(cloudcommon.GrpcCreds(tlsConfig),
		grpc.KeepaliveParams(serverParams),
		grpc.KeepaliveEnforcementPolicy(serverEnforcement),
		grpc.UnaryInterceptor(cloudcommon.AuditUnaryInterceptor),
		grpc.StreamInterceptor(cloudcommon.AuditStreamInterceptor),
	)
	edgeproto.RegisterNotifyApiServer(mgr.serv, mgr)
	if mgr.regServ != nil {
		mgr.regServ(mgr.serv)
	}
	log.DebugLog(log.DebugLevelNotify, "ServerMgr listening", "addr", addr)
	go func() {
		err = mgr.serv.Serve(lis)
		if err != nil {
			log.FatalLog("ServerMgr serve failed", "err", err)
		}
	}()
}

func (mgr *ServerMgr) Stop() {
	mgr.mux.Lock()
	mgr.serv.Stop()
	table := mgr.table
	mgr.table = nil
	mgr.mux.Unlock()
	if table != nil {
		for _, server := range table {
			server.Stop()
		}
	}
}

func (mgr *ServerMgr) StreamNotice(stream edgeproto.NotifyApi_StreamNoticeServer) error {
	ctx := stream.Context()
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return errors.New("Notify ServerMgr unable to get peer context")
	}
	peerAddr := peer.Addr.String()

	span := log.StartSpan(log.DebugLevelNotify, "StreamNotice start")
	span.SetTag("peer", peerAddr)
	spctx := log.ContextWithSpan(ctx, span)

	server := Server{}
	server.peerAddr = peerAddr
	server.running = make(chan struct{})
	server.sendrecv.init("server")
	server.sendrecv.peerAddr = peerAddr

	mgr.mux.Lock()
	server.notifyId = mgr.notifyId
	mgr.notifyId++
	for _, sendMany := range mgr.sends {
		send := sendMany.NewSend(peerAddr, server.notifyId)
		server.sendrecv.registerSend(send)
	}
	for _, recvMany := range mgr.recvs {
		recv := recvMany.NewRecv()
		server.sendrecv.registerRecv(recv)
	}
	mgr.mux.Unlock()

	// do initial version exchange
	err := server.negotiate(spctx, stream, mgr.name)
	if err != nil {
		server.logDisconnect(spctx, err)
		close(server.running)
		span.Finish()
		return err
	}

	// register server by client addr
	mgr.mux.Lock()
	mgr.table[peerAddr] = &server
	mgr.mux.Unlock()

	span.Finish()

	server.sendrecv.started = true
	server.sendrecv.sendRunning = make(chan struct{})
	server.sendrecv.recvRunning = make(chan struct{})
	// start send/recv threads.
	// recv thread will exit once stream is terminated after this
	// function returns.
	go server.sendrecv.recv(stream, server.notifyId, CleanupFlush)
	// to reduce number of threads, send is run inline
	server.sendrecv.send(stream)
	<-server.sendrecv.recvRunning

	span = log.StartSpan(log.DebugLevelNotify, "StreamNotice done")
	spctx = log.ContextWithSpan(ctx, span)
	server.logDisconnect(spctx, stream.Context().Err())

	mgr.mux.Lock()
	for ii, _ := range mgr.sends {
		mgr.sends[ii].DoneSend(peerAddr, server.sendrecv.sendlist[ii])
	}
	// another connection may come in from the same client so do not
	// remove it unless it's the same one.
	if remove, _ := mgr.table[peerAddr]; remove == &server {
		delete(mgr.table, peerAddr)
	}
	mgr.mux.Unlock()

	// Flush cache of objects from this connection.
	// NotifyId is used to identify objects sent by the lost connection.
	// The same client may reconnect, or the same objects may be resent
	// via a different intermediate node, but if they are updated, they
	// will have a new NotifyId, so will not be flushed.
	log.DebugLog(log.DebugLevelNotify, "Flush", "notifyId", server.notifyId)
	for _, recvMany := range mgr.recvs {
		recvMany.Flush(spctx, server.notifyId)
	}

	close(server.running)
	span.Finish()
	return err
}

func (mgr *ServerMgr) GetStats(peerAddr string) *Stats {
	stats := &Stats{}
	mgr.mux.Lock()
	defer mgr.mux.Unlock()
	if peerAddr != "" {
		server, found := mgr.table[peerAddr]
		if found {
			*stats = server.sendrecv.stats
			server.sendrecv.setObjStats(stats)
		}
	}
	return stats
}

// Get order of sends based on SendAll type
func (s *ServerMgr) GetSendOrder() map[reflect.Type]int {
	order := make(map[reflect.Type]int)
	for ii, send := range s.sends {
		order[reflect.TypeOf(send)] = ii
	}
	return order
}

func (s *Server) negotiate(ctx context.Context, stream edgeproto.NotifyApi_StreamNoticeServer, name string) error {
	var notice edgeproto.Notice
	// initial connection is version exchange
	// this also sets the connection Id so we can ignore spurious old
	// buffered messages
	req, err := stream.Recv()
	if err != nil {
		s.sendrecv.stats.NegotiateErrors++
		return err
	}
	if req.Action != edgeproto.NoticeAction_VERSION {
		log.DebugLog(log.DebugLevelNotify, "Notify server bad action", "expected", edgeproto.NoticeAction_VERSION, "got", notice.Action)
		s.sendrecv.stats.NegotiateErrors++
		return errors.New("Notify server expected action version")
	}
	s.sendrecv.setRemoteWanted(req.WantObjs)
	s.sendrecv.filterCloudletKeys = req.FilterCloudletKey
	if s.sendrecv.filterCloudletKeys {
		s.sendrecv.sendAllEnd = false
		s.sendrecv.manualSendAllEnd = true
	} else {
		s.sendrecv.sendAllEnd = true
	}
	// use lowest common version
	if req.Version > NotifyVersion {
		s.version = req.Version
	} else {
		s.version = NotifyVersion
	}
	if req.Tags != nil {
		if peer, found := req.Tags["name"]; found {
			s.sendrecv.peer = peer
		}
	}

	// send back my version
	notice.Action = edgeproto.NoticeAction_VERSION
	notice.Version = s.version
	notice.WantObjs = s.sendrecv.localWanted
	notice.Tags = map[string]string{
		"name": name,
	}
	err = stream.Send(&notice)
	if err != nil {
		s.sendrecv.stats.NegotiateErrors++
		return err
	}
	log.SpanLog(ctx, log.DebugLevelNotify, "Notify server connected",
		"client", s.peerAddr, "peer", s.sendrecv.peer, "version", s.version,
		"supported-version", NotifyVersion,
		"notifyid", s.notifyId,
		"remoteWanted", s.sendrecv.remoteWanted,
		"filterCloudletKey", s.sendrecv.filterCloudletKeys)
	return nil
}

func (s *Server) Stop() {
	s.mux.Lock()
	s.sendrecv.done = true
	s.sendrecv.wakeup()
	s.mux.Unlock()
	<-s.running
}

func (s *Server) logDisconnect(ctx context.Context, err error) {
	st, ok := status.FromError(err)
	if err == context.Canceled || (ok && st.Code() == codes.Canceled || err == nil) {
		log.SpanLog(ctx, log.DebugLevelNotify, "Notify server connection closed",
			"client", s.peerAddr, "peer", s.sendrecv.peer, "err", err)
	} else {
		log.SpanLog(ctx, log.DebugLevelInfo, "Notify server connection failed",
			"client", s.peerAddr, "peer", s.sendrecv.peer, "err", err)
	}
}
