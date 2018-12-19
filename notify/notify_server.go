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
	"net"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var NotifyRetryTime time.Duration = 250 * time.Millisecond

const NotifyVersion uint32 = 1

type SendAppHandler interface {
	GetAllKeys(keys map[edgeproto.AppKey]struct{})
	Get(key *edgeproto.AppKey, buf *edgeproto.App) bool
}

type SendAppInstHandler interface {
	GetAllKeys(keys map[edgeproto.AppInstKey]struct{})
	Get(key *edgeproto.AppInstKey, buf *edgeproto.AppInst) bool
	GetAppInstsForCloudlets(cloudlets map[edgeproto.CloudletKey]struct{},
		appInsts map[edgeproto.AppInstKey]struct{})
}

type SendCloudletHandler interface {
	GetAllKeys(keys map[edgeproto.CloudletKey]struct{})
	Get(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool
}

type SendFlavorHandler interface {
	GetAllKeys(keys map[edgeproto.FlavorKey]struct{})
	Get(key *edgeproto.FlavorKey, buf *edgeproto.Flavor) bool
}

type SendClusterFlavorHandler interface {
	GetAllKeys(keys map[edgeproto.ClusterFlavorKey]struct{})
	Get(key *edgeproto.ClusterFlavorKey, buf *edgeproto.ClusterFlavor) bool
}

type SendClusterInstHandler interface {
	Get(key *edgeproto.ClusterInstKey, buf *edgeproto.ClusterInst) bool
	GetClusterInstsForCloudlets(cloudlets map[edgeproto.CloudletKey]struct{},
		clusterInsts map[edgeproto.ClusterInstKey]struct{})
}

type RecvAppInstInfoHandler interface {
	Update(in *edgeproto.AppInstInfo, rev int64)
	Delete(in *edgeproto.AppInstInfo, rev int64)
	Flush(notifyId int64)
}

type RecvCloudletInfoHandler interface {
	Update(in *edgeproto.CloudletInfo, rev int64)
	Delete(in *edgeproto.CloudletInfo, rev int64)
	Flush(notifyId int64)
}

type RecvClusterInstInfoHandler interface {
	Update(in *edgeproto.ClusterInstInfo, rev int64)
	Delete(in *edgeproto.ClusterInstInfo, rev int64)
	Flush(notifyId int64)
}

type RecvMetricHandler interface {
	Recv(metric *edgeproto.Metric)
}

type RecvNodeHandler interface {
	Update(in *edgeproto.Node, rev int64)
	Delete(in *edgeproto.Node, rev int64)
	Flush(notifyId int64)
}

type ServerHandler interface {
	SendAppHandler() SendAppHandler
	SendAppInstHandler() SendAppInstHandler
	SendCloudletHandler() SendCloudletHandler
	SendFlavorHandler() SendFlavorHandler
	SendClusterFlavorHandler() SendClusterFlavorHandler
	SendClusterInstHandler() SendClusterInstHandler
	RecvAppInstInfoHandler() RecvAppInstInfoHandler
	RecvCloudletInfoHandler() RecvCloudletInfoHandler
	RecvClusterInstInfoHandler() RecvClusterInstInfoHandler
	RecvMetricHandler() RecvMetricHandler
	RecvNodeHandler() RecvNodeHandler
}

type ServerStats struct {
	AppsSent           uint64
	AppInstsSent       uint64
	CloudletsSent      uint64
	FlavorsSent        uint64
	ClusterFlavorsSent uint64
	ClusterInstsSent   uint64
	NegotiateErrors    uint64
	RecvErrors         uint64
	SendErrors         uint64
}

// Server is on the upstream side and sends data to downstream clients.
// On first connect, it will send all data from the database that is
// required by the client. After that it will send objects only when
// they are changed.
type Server struct {
	peerAddr       string
	apps           map[edgeproto.AppKey]struct{}
	appInsts       map[edgeproto.AppInstKey]struct{}
	cloudlets      map[edgeproto.CloudletKey]struct{}
	flavors        map[edgeproto.FlavorKey]struct{}
	clusterflavors map[edgeproto.ClusterFlavorKey]struct{}
	clusterInsts   map[edgeproto.ClusterInstKey]struct{}
	mux            util.Mutex
	signal         chan bool
	done           bool
	handler        ServerHandler
	stats          ServerStats
	version        uint32
	notifyId       int64
	requestor      edgeproto.NoticeRequestor
	running        chan struct{}
	// tracked cloudlets are the set of keys received from the client
	// for requestor type CRM. We only send data related to these
	// cloudlet key(s)
	trackedCloudlets map[edgeproto.CloudletKey]struct{}
}

// ServerMgr maintains all the Server threads for clients connected to us.
type ServerMgr struct {
	table    map[string]*Server
	mux      util.Mutex
	handler  ServerHandler
	notifyId int64
	serv     *grpc.Server
}

var ServerMgrOne ServerMgr

func (mgr *ServerMgr) Start(addr string, tlsCertFile string, handler ServerHandler) {
	mgr.mux.Lock()
	defer mgr.mux.Unlock()

	if mgr.table != nil {
		return
	}
	mgr.table = make(map[string]*Server)
	mgr.handler = handler

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.FatalLog("ServerMgr listen failed", "err", err)
	}

	creds, err := tls.GetTLSServerCreds(tlsCertFile)
	if err != nil {
		log.FatalLog("Failed to get TLS creds", "err", err)
	}
	mgr.serv = grpc.NewServer(grpc.Creds(creds), grpc.KeepaliveParams(serverParams), grpc.KeepaliveEnforcementPolicy(serverEnforcement))
	edgeproto.RegisterNotifyApiServer(mgr.serv, mgr)
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
	mgr.handler = nil
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

	server := Server{}
	server.peerAddr = peerAddr
	server.apps = make(map[edgeproto.AppKey]struct{})
	server.appInsts = make(map[edgeproto.AppInstKey]struct{})
	server.cloudlets = make(map[edgeproto.CloudletKey]struct{})
	server.flavors = make(map[edgeproto.FlavorKey]struct{})
	server.clusterflavors = make(map[edgeproto.ClusterFlavorKey]struct{})
	server.clusterInsts = make(map[edgeproto.ClusterInstKey]struct{})
	server.signal = make(chan bool, 1)
	server.handler = mgr.handler
	server.running = make(chan struct{})
	server.trackedCloudlets = make(map[edgeproto.CloudletKey]struct{})

	// do initial version exchange
	err := server.negotiate(stream)
	if err != nil {
		server.logDisconnect(err)
		close(server.running)
		return err
	}

	// register server by client addr
	mgr.mux.Lock()
	mgr.table[peerAddr] = &server
	server.notifyId = mgr.notifyId
	mgr.notifyId++
	mgr.mux.Unlock()

	// start send/recv threads.
	// recv thread will exit once stream is terminated after this
	// function returns.
	go server.recv(stream)
	// to reduce number of threads, send is run inline
	server.send(stream)
	server.logDisconnect(stream.Context().Err())

	mgr.mux.Lock()
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
	recvAppInstInfo := server.handler.RecvAppInstInfoHandler()
	if recvAppInstInfo != nil {
		recvAppInstInfo.Flush(server.notifyId)
	}
	recvCloudletInfo := server.handler.RecvCloudletInfoHandler()
	if recvCloudletInfo != nil {
		recvCloudletInfo.Flush(server.notifyId)
	}
	recvClusterInstInfo := server.handler.RecvClusterInstInfoHandler()
	if recvClusterInstInfo != nil {
		recvClusterInstInfo.Flush(server.notifyId)
	}
	recvNode := server.handler.RecvNodeHandler()
	if recvNode != nil {
		recvNode.Flush(server.notifyId)
	}

	close(server.running)
	return err
}

func (mgr *ServerMgr) GetStats(peerAddr string) *ServerStats {
	stats := &ServerStats{}
	mgr.mux.Lock()
	defer mgr.mux.Unlock()
	if peerAddr != "" {
		server, found := mgr.table[peerAddr]
		if found {
			*stats = server.stats
		}
	}
	return stats
}

func (mgr *ServerMgr) UpdateApp(key *edgeproto.AppKey, old *edgeproto.App) {
	mgr.mux.Lock()
	defer mgr.mux.Unlock()
	for _, server := range mgr.table {
		if server.requestor != edgeproto.NoticeRequestor_NoticeRequestorDME {
			// Apps to crm triggered by AppInst update
			continue
		}

		server.UpdateApp(key)
	}
}

func (mgr *ServerMgr) UpdateAppInst(key *edgeproto.AppInstKey, old *edgeproto.AppInst) {
	mgr.mux.Lock()
	defer mgr.mux.Unlock()
	for _, server := range mgr.table {
		server.UpdateAppInst(key)
	}
}

func (mgr *ServerMgr) UpdateCloudlet(key *edgeproto.CloudletKey, old *edgeproto.Cloudlet) {
	mgr.mux.Lock()
	defer mgr.mux.Unlock()
	for _, server := range mgr.table {
		if server.requestor == edgeproto.NoticeRequestor_NoticeRequestorMEXInfra {
			log.DebugLog(log.DebugLevelNotify,
				"Updating notify to send the cluster info for MEXINFRA",
				"cloudlet", key.Name)

			server.updateTrackedCloudlets(key, register)
		}
		if server.requestor != edgeproto.NoticeRequestor_NoticeRequestorCRM {
			continue
		}
		server.UpdateCloudlet(key)
	}
}

func (mgr *ServerMgr) UpdateFlavor(key *edgeproto.FlavorKey, old *edgeproto.Flavor) {
	mgr.mux.Lock()
	defer mgr.mux.Unlock()
	for _, server := range mgr.table {
		if server.requestor != edgeproto.NoticeRequestor_NoticeRequestorCRM {
			continue
		}
		server.UpdateFlavor(key)
	}
}

func (mgr *ServerMgr) UpdateClusterFlavor(key *edgeproto.ClusterFlavorKey, old *edgeproto.ClusterFlavor) {
	mgr.mux.Lock()
	defer mgr.mux.Unlock()
	for _, server := range mgr.table {
		if server.requestor != edgeproto.NoticeRequestor_NoticeRequestorCRM {
			continue
		}
		server.UpdateClusterFlavor(key)
	}
}

func (mgr *ServerMgr) UpdateClusterInst(key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInst) {
	mgr.mux.Lock()
	defer mgr.mux.Unlock()
	for _, server := range mgr.table {
		if server.requestor != edgeproto.NoticeRequestor_NoticeRequestorCRM &&
			server.requestor != edgeproto.NoticeRequestor_NoticeRequestorMEXInfra {
			continue
		}
		server.UpdateClusterInst(key)
	}
}

func (s *Server) wakeup() {
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

func (s *Server) UpdateApp(key *edgeproto.AppKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.apps[*key] = struct{}{}
	s.wakeup()
}

func (s *Server) UpdateAppInst(key *edgeproto.AppInstKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.requestor == edgeproto.NoticeRequestor_NoticeRequestorCRM {
		if _, found := s.trackedCloudlets[key.CloudletKey]; !found {
			// not tracked by this client
			return
		}
		// for crm, also trigger sending app
		s.apps[key.AppKey] = struct{}{}
	}
	s.appInsts[*key] = struct{}{}
	s.wakeup()
}

func (s *Server) UpdateCloudlet(key *edgeproto.CloudletKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, found := s.trackedCloudlets[*key]; !found {
		// not tracked by this client
		return
	}
	s.cloudlets[*key] = struct{}{}
	s.wakeup()
}

func (s *Server) UpdateFlavor(key *edgeproto.FlavorKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.flavors[*key] = struct{}{}
	s.wakeup()
}

func (s *Server) UpdateClusterFlavor(key *edgeproto.ClusterFlavorKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.clusterflavors[*key] = struct{}{}
	s.wakeup()
}

func (s *Server) UpdateClusterInst(key *edgeproto.ClusterInstKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, found := s.trackedCloudlets[key.CloudletKey]; !found {
		// not tracked by this client
		return
	}
	s.clusterInsts[*key] = struct{}{}
	s.wakeup()
}

func (s *Server) negotiate(stream edgeproto.NotifyApi_StreamNoticeServer) error {
	var notice edgeproto.NoticeReply
	// initial connection is version exchange
	// this also sets the connection Id so we can ignore spurious old
	// buffered messages
	req, err := stream.Recv()
	if err != nil {
		s.stats.NegotiateErrors++
		return err
	}
	if req.Action != edgeproto.NoticeAction_VERSION {
		log.DebugLog(log.DebugLevelNotify, "Notify server bad action", "expected", edgeproto.NoticeAction_VERSION, "got", notice.Action)
		s.stats.NegotiateErrors++
		return errors.New("Notify server expected action version")
	}

	if req.Requestor != edgeproto.NoticeRequestor_NoticeRequestorDME && req.Requestor != edgeproto.NoticeRequestor_NoticeRequestorCRM &&
		req.Requestor != edgeproto.NoticeRequestor_NoticeRequestorMEXInfra {
		s.stats.NegotiateErrors++
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
		s.stats.NegotiateErrors++
		return err
	}
	log.DebugLog(log.DebugLevelNotify, "Notify server connected",
		"client", s.peerAddr, "version", s.version,
		"supported-version", NotifyVersion)
	return nil
}

func (s *Server) send(stream edgeproto.NotifyApi_StreamNoticeServer) {
	var err error
	var notice edgeproto.NoticeReply
	var noticeApp edgeproto.NoticeReply_App
	var noticeAppInst edgeproto.NoticeReply_AppInst
	var noticeCloudlet edgeproto.NoticeReply_Cloudlet
	var noticeFlavor edgeproto.NoticeReply_Flavor
	var noticeClusterFlavor edgeproto.NoticeReply_ClusterFlavor
	var noticeClusterInst edgeproto.NoticeReply_ClusterInst
	var app edgeproto.App
	var appInst edgeproto.AppInst
	var cloudlet edgeproto.Cloudlet
	var flavor edgeproto.Flavor
	var clusterflavor edgeproto.ClusterFlavor
	var clusterInst edgeproto.ClusterInst

	noticeApp.App = &app
	noticeAppInst.AppInst = &appInst
	noticeCloudlet.Cloudlet = &cloudlet
	noticeFlavor.Flavor = &flavor
	noticeClusterFlavor.ClusterFlavor = &clusterflavor
	noticeClusterInst.ClusterInst = &clusterInst
	sendAll := true
	sendApp := s.handler.SendAppHandler()
	sendAppInst := s.handler.SendAppInstHandler()
	sendCloudlet := s.handler.SendCloudletHandler()
	sendFlavor := s.handler.SendFlavorHandler()
	sendClusterFlavor := s.handler.SendClusterFlavorHandler()
	sendClusterInst := s.handler.SendClusterInstHandler()
	// trigger initial sendAll
	s.wakeup()

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
			s.done = true
		}
		s.mux.Lock()
		if len(s.appInsts) == 0 && len(s.cloudlets) == 0 &&
			len(s.flavors) == 0 && len(s.clusterInsts) == 0 &&
			len(s.clusterflavors) == 0 && len(s.apps) == 0 &&
			!s.done && !sendAll && stream.Context().Err() == nil {
			s.mux.Unlock()
			continue
		}
		apps := s.apps
		appInsts := s.appInsts
		cloudlets := s.cloudlets
		flavors := s.flavors
		clusterflavors := s.clusterflavors
		clusterInsts := s.clusterInsts
		s.apps = make(map[edgeproto.AppKey]struct{})
		s.appInsts = make(map[edgeproto.AppInstKey]struct{})
		s.cloudlets = make(map[edgeproto.CloudletKey]struct{})
		s.flavors = make(map[edgeproto.FlavorKey]struct{})
		s.clusterflavors = make(map[edgeproto.ClusterFlavorKey]struct{})
		s.clusterInsts = make(map[edgeproto.ClusterInstKey]struct{})
		s.mux.Unlock()
		if s.done {
			break
		}
		if sendAll {
			log.DebugLog(log.DebugLevelNotify, "Send all",
				"client", s.peerAddr,
				"requestor", s.requestor)
			if s.requestor == edgeproto.NoticeRequestor_NoticeRequestorCRM {
				if sendFlavor != nil {
					sendFlavor.GetAllKeys(flavors)
				}
				if sendClusterFlavor != nil {
					sendClusterFlavor.GetAllKeys(clusterflavors)
				}
				// Cloudlet, AppInsts, CloudletInsts sends are
				// triggered when registering a new CloudletInfo
				// on receive, so there is no need to send all here.
			} else if s.requestor == edgeproto.NoticeRequestor_NoticeRequestorMEXInfra {
				// need to update the interested cloudlets
				sendCloudlet.GetAllKeys(cloudlets)
				for k, _ := range cloudlets {
					s.updateTrackedCloudlets(&k, register)
				}
				if sendClusterInst != nil {
					sendClusterInst.GetClusterInstsForCloudlets(cloudlets, clusterInsts)
				}
			} else {
				if sendApp != nil {
					sendApp.GetAllKeys(apps)
					log.DebugLog(log.DebugLevelNotify, "all apps", "count", len(apps))
				}
				if sendAppInst != nil {
					sendAppInst.GetAllKeys(appInsts)
					log.DebugLog(log.DebugLevelNotify, "all AppInsts", "count", len(appInsts))
				}
			}
		}

		if sendFlavor != nil {
			// send flavors
			notice.Data = &noticeFlavor
			for key, _ := range flavors {
				found := sendFlavor.Get(&key, &flavor)
				if found {
					notice.Action = edgeproto.NoticeAction_UPDATE
				} else {
					notice.Action = edgeproto.NoticeAction_DELETE
					flavor.Key = key
				}
				log.DebugLog(log.DebugLevelNotify, "Send Flavor",
					"client", s.peerAddr,
					"action", notice.Action,
					"key", flavor.Key.GetKeyString())
				err = stream.Send(&notice)
				if err != nil {
					break
				}
				s.stats.FlavorsSent++
			}
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		if sendClusterFlavor != nil {
			// send cluster flavors
			notice.Data = &noticeClusterFlavor
			for key, _ := range clusterflavors {
				found := sendClusterFlavor.Get(&key, &clusterflavor)
				if found {
					notice.Action = edgeproto.NoticeAction_UPDATE
				} else {
					notice.Action = edgeproto.NoticeAction_DELETE
					clusterflavor.Key = key
				}
				log.DebugLog(log.DebugLevelNotify, "Send ClusterFlavor",
					"client", s.peerAddr,
					"action", notice.Action,
					"key", clusterflavor.Key.GetKeyString())
				err = stream.Send(&notice)
				if err != nil {
					break
				}
				s.stats.ClusterFlavorsSent++
			}
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		if sendClusterInst != nil {
			// send clusterInsts
			notice.Data = &noticeClusterInst
			for key, _ := range clusterInsts {
				found := sendClusterInst.Get(&key, &clusterInst)
				if found {
					notice.Action = edgeproto.NoticeAction_UPDATE
				} else {
					notice.Action = edgeproto.NoticeAction_DELETE
					clusterInst.Key = key
				}
				log.DebugLog(log.DebugLevelNotify, "Send ClusterInst",
					"client", s.peerAddr,
					"action", notice.Action,
					"key", clusterInst.Key.GetKeyString())
				err = stream.Send(&notice)
				if err != nil {
					break
				}
				s.stats.ClusterInstsSent++
			}
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		if sendApp != nil {
			// send apps
			notice.Data = &noticeApp
			for key, _ := range apps {
				found := sendApp.Get(&key, &app)
				if found {
					notice.Action = edgeproto.NoticeAction_UPDATE
				} else {
					notice.Action = edgeproto.NoticeAction_DELETE
					app.Key = key
				}
				log.DebugLog(log.DebugLevelNotify, "Send App",
					"client", s.peerAddr,
					"action", notice.Action,
					"key", app.Key.GetKeyString())
				err = stream.Send(&notice)
				if err != nil {
					break
				}
				s.stats.AppsSent++
			}
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		if sendAppInst != nil {
			// send appInsts
			notice.Data = &noticeAppInst
			for key, _ := range appInsts {
				found := sendAppInst.Get(&key, &appInst)
				if found {
					notice.Action = edgeproto.NoticeAction_UPDATE
				} else {
					notice.Action = edgeproto.NoticeAction_DELETE
					appInst.Key = key
				}
				log.DebugLog(log.DebugLevelNotify, "Send AppInst",
					"client", s.peerAddr,
					"action", notice.Action,
					"key", appInst.Key.GetKeyString())
				err = stream.Send(&notice)
				if err != nil {
					break
				}
				s.stats.AppInstsSent++
			}
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		if sendCloudlet != nil {
			// send cloudlets
			notice.Data = &noticeCloudlet
			for key, _ := range cloudlets {
				found := sendCloudlet.Get(&key, &cloudlet)
				if found {
					notice.Action = edgeproto.NoticeAction_UPDATE
				} else {
					notice.Action = edgeproto.NoticeAction_DELETE
					cloudlet.Key = key
				}
				log.DebugLog(log.DebugLevelNotify, "Send Cloudlet",
					"client", s.peerAddr,
					"action", notice.Action,
					"key", cloudlet.Key.GetKeyString())
				err = stream.Send(&notice)
				if err != nil {
					break
				}
				s.stats.CloudletsSent++
			}
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		if sendAll {
			notice.Action = edgeproto.NoticeAction_SENDALL_END
			notice.Data = nil
			err = stream.Send(&notice)
			if err != nil {
				s.stats.SendErrors++
				break
			}
			sendAll = false
		}
		apps = nil
		appInsts = nil
		cloudlets = nil
		flavors = nil
		clusterflavors = nil
		clusterInsts = nil
	}
}

func (s *Server) recv(stream edgeproto.NotifyApi_StreamNoticeServer) {
	recvAppInstInfo := s.handler.RecvAppInstInfoHandler()
	recvCloudletInfo := s.handler.RecvCloudletInfoHandler()
	recvClusterInstInfo := s.handler.RecvClusterInstInfoHandler()
	recvMetric := s.handler.RecvMetricHandler()
	recvNode := s.handler.RecvNodeHandler()
	for !s.done {
		req, err := stream.Recv()
		if s.done {
			break
		}
		if err != nil {
			log.DebugLog(log.DebugLevelNotify, "Server receive", "err", err)
			break
		}
		if req.Action != edgeproto.NoticeAction_UPDATE &&
			req.Action != edgeproto.NoticeAction_DELETE {
			log.DebugLog(log.DebugLevelNotify,
				"Server recv unexpected notice action",
				"action", req.Action)
			s.stats.RecvErrors++
			continue
		}
		appInstInfo := req.GetAppInstInfo()
		if recvAppInstInfo != nil && appInstInfo != nil {
			appInstInfo.NotifyId = s.notifyId
			log.DebugLog(log.DebugLevelNotify, "Recv app inst info",
				"client", s.peerAddr,
				"action", req.Action,
				"info", appInstInfo)
			if req.Action == edgeproto.NoticeAction_UPDATE {
				recvAppInstInfo.Update(appInstInfo, s.notifyId)
			} else if req.Action == edgeproto.NoticeAction_DELETE {
				recvAppInstInfo.Delete(appInstInfo, s.notifyId)
			}
		}
		clusterInstInfo := req.GetClusterInstInfo()
		if recvClusterInstInfo != nil && clusterInstInfo != nil {
			clusterInstInfo.NotifyId = s.notifyId
			log.DebugLog(log.DebugLevelNotify, "Recv cluster inst info",
				"client", s.peerAddr,
				"action", req.Action,
				"info", clusterInstInfo)
			if req.Action == edgeproto.NoticeAction_UPDATE {
				recvClusterInstInfo.Update(clusterInstInfo, s.notifyId)
			} else if req.Action == edgeproto.NoticeAction_DELETE {
				recvClusterInstInfo.Delete(clusterInstInfo, s.notifyId)
			}
		}
		cloudletInfo := req.GetCloudletInfo()
		if recvCloudletInfo != nil && cloudletInfo != nil {
			cloudletInfo.NotifyId = s.notifyId
			log.DebugLog(log.DebugLevelNotify, "Recv cloudlet info",
				"client", s.peerAddr,
				"action", req.Action,
				"info", cloudletInfo)
			if req.Action == edgeproto.NoticeAction_UPDATE {
				recvCloudletInfo.Update(cloudletInfo, s.notifyId)
				s.updateTrackedCloudlets(&cloudletInfo.Key, register)
			} else if req.Action == edgeproto.NoticeAction_DELETE {
				recvCloudletInfo.Delete(cloudletInfo, s.notifyId)
				s.updateTrackedCloudlets(&cloudletInfo.Key, unregister)
			}
		}
		metric := req.GetMetric()
		if recvMetric != nil && metric != nil {
			recvMetric.Recv(metric)
		}
		node := req.GetNode()
		if recvNode != nil && node != nil {
			node.NotifyId = s.notifyId
			log.DebugLog(log.DebugLevelNotify, "Recv node",
				"client", s.peerAddr,
				"action", req.Action,
				"node", node)
			if req.Action == edgeproto.NoticeAction_UPDATE {
				recvNode.Update(node, s.notifyId)
			} else if req.Action == edgeproto.NoticeAction_DELETE {
				recvNode.Delete(node, s.notifyId)
			}
		}
	}
}

func (s *Server) Stop() {
	s.mux.Lock()
	s.done = true
	s.wakeup()
	s.mux.Unlock()
	<-s.running
}

func (s *Server) logDisconnect(err error) {
	st, ok := status.FromError(err)
	if err == context.Canceled || (ok && st.Code() == codes.Canceled || err == nil) {
		log.DebugLog(log.DebugLevelNotify, "Notify server connection closed",
			"client", s.peerAddr, "err", err)
	} else {
		log.InfoLog("Notify server connection failed",
			"client", s.peerAddr, "err", err)
	}
}

type registerAction int

const (
	register registerAction = iota
	unregister
)

func (s *Server) updateTrackedCloudlets(key *edgeproto.CloudletKey, action registerAction) {
	s.mux.Lock()
	if action == register {
		s.trackedCloudlets[*key] = struct{}{}
	} else {
		delete(s.trackedCloudlets, *key)
	}
	s.mux.Unlock()

	if s.requestor != edgeproto.NoticeRequestor_NoticeRequestorCRM {
		// DMEs also send cloudletInfo but it's just to register
		// the tracked cloudlet key.
		return
	}

	cloudlets := make(map[edgeproto.CloudletKey]struct{})
	cloudlets[*key] = struct{}{}
	appInsts := make(map[edgeproto.AppInstKey]struct{})
	clusterInsts := make(map[edgeproto.ClusterInstKey]struct{})

	sendAppInst := s.handler.SendAppInstHandler()
	if sendAppInst != nil {
		sendAppInst.GetAppInstsForCloudlets(cloudlets, appInsts)
	}
	sendClusterInst := s.handler.SendClusterInstHandler()
	if sendClusterInst != nil {
		sendClusterInst.GetClusterInstsForCloudlets(cloudlets, clusterInsts)
	}

	if action == register {
		// trigger sends of all objects related to cloudlet
		s.UpdateCloudlet(key)
		for k, _ := range clusterInsts {
			s.UpdateClusterInst(&k)
		}
		for k, _ := range appInsts {
			s.UpdateAppInst(&k)
		}
	} else {
		// TODO:
		// trigger deletes for objects no longer tracked
		// This helps clean up unnecessary objects, but is really
		// only needed for intermediate cache nodes between
		// the controller and multiple CRMs.
	}
}
