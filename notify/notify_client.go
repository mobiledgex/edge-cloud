package notify

import (
	"context"
	"math/rand"
	"net"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

const MetricsLimit = 500

type Client struct {
	addrs       []string
	tlsCertFile string
	handler     ClientHandler
	requestor   edgeproto.NoticeRequestor
	version     uint32
	stats       ClientStats
	addrIdx     int
	mux         util.Mutex
	signal      chan bool
	done        bool
	conn        *grpc.ClientConn
	cancel      context.CancelFunc
	running     chan struct{}
	sendRunning chan struct{}
	recvRunning chan struct{}
	// localAddr has its own lock because its called in some
	// grpc interceptor context
	localAddr    string
	localAddrMux util.Mutex
	// The following fields are for the send thread
	appInstInfos     map[edgeproto.AppInstKey]struct{}
	clusterInstInfos map[edgeproto.ClusterInstKey]struct{}
	cloudletInfos    map[edgeproto.CloudletKey]struct{}
	nodes            map[edgeproto.NodeKey]struct{}
	metrics          []*edgeproto.Metric
}

type ClientRecvAllMaps struct {
	AppInsts       map[edgeproto.AppInstKey]struct{}
	Cloudlets      map[edgeproto.CloudletKey]struct{}
	Flavors        map[edgeproto.FlavorKey]struct{}
	ClusterFlavors map[edgeproto.ClusterFlavorKey]struct{}
}

type SendAppInstInfoHandler interface {
	GetAllKeys(keys map[edgeproto.AppInstKey]struct{})
	Get(key *edgeproto.AppInstKey, buf *edgeproto.AppInstInfo) bool
}

type SendCloudletInfoHandler interface {
	GetAllKeys(keys map[edgeproto.CloudletKey]struct{})
	Get(key *edgeproto.CloudletKey, buf *edgeproto.CloudletInfo) bool
}

type SendClusterInstInfoHandler interface {
	GetAllKeys(keys map[edgeproto.ClusterInstKey]struct{})
	Get(key *edgeproto.ClusterInstKey, buf *edgeproto.ClusterInstInfo) bool
}

type SendNodeHandler interface {
	GetAllKeys(keys map[edgeproto.NodeKey]struct{})
	Get(key *edgeproto.NodeKey, buf *edgeproto.Node) bool
}

type RecvAppInstHandler interface {
	Update(in *edgeproto.AppInst, rev int64)
	Delete(in *edgeproto.AppInst, rev int64)
	Prune(keys map[edgeproto.AppInstKey]struct{})
}

type RecvCloudletHandler interface {
	Update(in *edgeproto.Cloudlet, rev int64)
	Delete(in *edgeproto.Cloudlet, rev int64)
	Prune(keys map[edgeproto.CloudletKey]struct{})
}

type RecvFlavorHandler interface {
	Update(in *edgeproto.Flavor, rev int64)
	Delete(in *edgeproto.Flavor, rev int64)
	Prune(keys map[edgeproto.FlavorKey]struct{})
}

type RecvClusterFlavorHandler interface {
	Update(in *edgeproto.ClusterFlavor, rev int64)
	Delete(in *edgeproto.ClusterFlavor, rev int64)
	Prune(keys map[edgeproto.ClusterFlavorKey]struct{})
}

type RecvClusterInstHandler interface {
	Update(in *edgeproto.ClusterInst, rev int64)
	Delete(in *edgeproto.ClusterInst, rev int64)
	Prune(keys map[edgeproto.ClusterInstKey]struct{})
}

type ClientHandler interface {
	SendAppInstInfoHandler() SendAppInstInfoHandler
	SendCloudletInfoHandler() SendCloudletInfoHandler
	SendClusterInstInfoHandler() SendClusterInstInfoHandler
	SendNodeHandler() SendNodeHandler
	RecvAppInstHandler() RecvAppInstHandler
	RecvCloudletHandler() RecvCloudletHandler
	RecvFlavorHandler() RecvFlavorHandler
	RecvClusterFlavorHandler() RecvClusterFlavorHandler
	RecvClusterInstHandler() RecvClusterInstHandler
}

type ClientStats struct {
	Tries                uint64
	Connects             uint64
	NegotiateErrors      uint64
	RecvErrors           uint64
	SendErrors           uint64
	AppInstInfosSent     uint64
	ClusterInstInfosSent uint64
	CloudletInfosSent    uint64
	NodesSent            uint64
	MetricsSent          uint64
	MetricsDropped       uint64
	AppInstRecv          uint64
	CloudletRecv         uint64
	FlavorRecv           uint64
	ClusterFlavorRecv    uint64
	ClusterInstRecv      uint64
	Recv                 uint64
}

func cancelNoop() {}

func NewDMEClient(addrs []string, tlsCertFile string, handler ClientHandler) *Client {
	return newClient(addrs, tlsCertFile, handler, edgeproto.NoticeRequestor_NoticeRequestorDME)
}

func NewCRMClient(addrs []string, tlsCertFile string, handler ClientHandler) *Client {
	return newClient(addrs, tlsCertFile, handler, edgeproto.NoticeRequestor_NoticeRequestorCRM)
}

func newClient(addrs []string, tlsCertFile string, handler ClientHandler, requestor edgeproto.NoticeRequestor) *Client {
	s := Client{}
	s.addrs = addrs
	s.tlsCertFile = tlsCertFile
	s.handler = handler
	s.requestor = requestor
	s.signal = make(chan bool, 1)
	s.appInstInfos = make(map[edgeproto.AppInstKey]struct{})
	s.clusterInstInfos = make(map[edgeproto.ClusterInstKey]struct{})
	s.cloudletInfos = make(map[edgeproto.CloudletKey]struct{})
	s.nodes = make(map[edgeproto.NodeKey]struct{})
	s.metrics = make([]*edgeproto.Metric, 0)
	return &s
}

func (s *Client) Start() {
	s.mux.Lock()
	s.done = false
	s.cancel = cancelNoop
	s.running = make(chan struct{})
	s.addrIdx = rand.Int() % len(s.addrs)
	s.mux.Unlock()
	go s.run()
}

func (s *Client) Stop() {
	s.mux.Lock()
	s.done = true
	s.cancel()
	s.mux.Unlock()
	<-s.running
}

func (s *Client) run() {
	for !s.done {
		// connect to server
		stream, err := s.connect()
		if err != nil {
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				time.Sleep(NotifyRetryTime)
			}
			continue
		}
		// do initial version exchange
		err = s.negotiate(stream)
		if err != nil {
			s.logDisconnect(err)
			s.connectCleanup()
			continue
		}

		s.sendRunning = make(chan struct{})
		s.recvRunning = make(chan struct{})
		go s.send(stream)
		go s.recv(stream)
		// if there is a communication error, both threads will exit
		<-s.sendRunning
		<-s.recvRunning
		s.logDisconnect(stream.Context().Err())
		s.connectCleanup()
	}
	log.DebugLog(log.DebugLevelNotify, "Notify client stopped",
		"server", s.GetServerAddr(), "local", s.GetLocalAddr())
	close(s.running)
}

// connect to the server
func (s *Client) connect() (edgeproto.NotifyApi_StreamNoticeClient, error) {
	// connect to server
	s.mux.Lock()
	s.stats.Tries++
	s.addrIdx++
	if s.addrIdx == len(s.addrs) {
		s.addrIdx = 0
	}
	addr := s.addrs[s.addrIdx]
	s.mux.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), NotifyRetryTime)
	dialOption, err := tls.GetTLSClientDialOption(addr, s.tlsCertFile)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.DialContext(ctx, addr,
		dialOption,
		grpc.WithStatsHandler(&grpcStatsHandler{client: s}),
		grpc.WithKeepaliveParams(clientParams))
	cancel()
	if err != nil {
		return nil, err
	}
	log.DebugLog(log.DebugLevelNotify, "creating notify client", "addr", addr, "tlsCert", s.tlsCertFile)

	api := edgeproto.NewNotifyApiClient(conn)
	ctx, cancel = context.WithCancel(context.Background())
	stream, err := api.StreamNotice(ctx)
	if err != nil {
		log.DebugLog(log.DebugLevelNotify, "Notify client get stream",
			"server", addr, "error", err)
		cancel()
		conn.Close()
		return nil, err
	}
	s.mux.Lock()
	s.cancel = cancel
	s.conn = conn
	s.stats.Connects++
	s.mux.Unlock()
	return stream, nil
}

func (s *Client) connectCleanup() {
	s.mux.Lock()
	s.conn.Close()
	s.cancel()
	s.cancel = cancelNoop
	s.mux.Unlock()
	s.SetLocalAddr("")
}

// negotiate performs the initial exchange between client and server
// that register what type of client we are with the server, and
// settle upon the maximum supported version.
func (s *Client) negotiate(stream edgeproto.NotifyApi_StreamNoticeClient) error {
	var request edgeproto.NoticeRequest
	var reply *edgeproto.NoticeReply
	// Send our version and read back remote version.
	// We use the lowest common version.
	request.Version = NotifyVersion
	request.Action = edgeproto.NoticeAction_VERSION
	request.Requestor = s.requestor
	err := stream.Send(&request)
	if err != nil {
		s.stats.NegotiateErrors++
		return err
	}
	reply, err = stream.Recv()
	if err != nil {
		s.stats.NegotiateErrors++
		return err
	}
	if request.Version > reply.Version {
		s.version = reply.Version
	} else {
		s.version = request.Version
	}
	s.mux.Lock()
	addr := s.addrs[s.addrIdx]
	s.mux.Unlock()
	log.DebugLog(log.DebugLevelNotify, "Notify client connected",
		"server", addr, "local", s.GetLocalAddr(),
		"version", s.version, "supported-version", NotifyVersion,
		"tries", s.stats.Tries, "connects", s.stats.Connects)
	return nil
}

func (s *Client) recv(stream edgeproto.NotifyApi_StreamNoticeClient) {
	// server will send all data first
	allMaps := &ClientRecvAllMaps{}
	allMaps.AppInsts = make(map[edgeproto.AppInstKey]struct{})
	allMaps.Cloudlets = make(map[edgeproto.CloudletKey]struct{})
	allMaps.Flavors = make(map[edgeproto.FlavorKey]struct{})
	allMaps.ClusterFlavors = make(map[edgeproto.ClusterFlavorKey]struct{})
	recvAppInst := s.handler.RecvAppInstHandler()
	recvCloudlet := s.handler.RecvCloudletHandler()
	recvFlavor := s.handler.RecvFlavorHandler()
	recvClusterFlavor := s.handler.RecvClusterFlavorHandler()
	recvClusterInst := s.handler.RecvClusterInstHandler()
	for !s.done {
		reply, err := stream.Recv()
		if s.done {
			break
		}
		if err != nil {
			break
		}
		if allMaps != nil && reply.Action == edgeproto.NoticeAction_UPDATE {
			appInst := reply.GetAppInst()
			if appInst != nil {
				allMaps.AppInsts[appInst.Key] = struct{}{}
			}
			cloudlet := reply.GetCloudlet()
			if cloudlet != nil {
				allMaps.Cloudlets[cloudlet.Key] = struct{}{}
			}
			flavor := reply.GetFlavor()
			if flavor != nil {
				allMaps.Flavors[flavor.Key] = struct{}{}
			}
			clusterflavor := reply.GetClusterFlavor()
			if clusterflavor != nil {
				allMaps.ClusterFlavors[clusterflavor.Key] = struct{}{}
			}
		}
		if reply.Action == edgeproto.NoticeAction_SENDALL_END {
			if recvAppInst != nil {
				recvAppInst.Prune(allMaps.AppInsts)
			}
			if recvCloudlet != nil {
				recvCloudlet.Prune(allMaps.Cloudlets)
			}
			if recvFlavor != nil {
				recvFlavor.Prune(allMaps.Flavors)
			}
			if recvClusterFlavor != nil {
				recvClusterFlavor.Prune(allMaps.ClusterFlavors)
			}
			allMaps = nil
			continue
		}
		if reply.Action != edgeproto.NoticeAction_UPDATE &&
			reply.Action != edgeproto.NoticeAction_DELETE {
			log.DebugLog(log.DebugLevelNotify,
				"Client recv unexpected notice action",
				"action", reply.Action)
			s.stats.RecvErrors++
			continue
		}
		flavor := reply.GetFlavor()
		if recvFlavor != nil && flavor != nil {
			if reply.Action == edgeproto.NoticeAction_UPDATE {
				log.DebugLog(log.DebugLevelNotify,
					"client flavor update",
					"key", flavor.Key.GetKeyString())
				recvFlavor.Update(flavor, 0)
			} else if reply.Action == edgeproto.NoticeAction_DELETE {
				log.DebugLog(log.DebugLevelNotify,
					"client flavor delete",
					"key", flavor.Key.GetKeyString())
				recvFlavor.Delete(flavor, 0)
			}
			s.stats.FlavorRecv++
			s.stats.Recv++
		}
		clusterflavor := reply.GetClusterFlavor()
		if recvClusterFlavor != nil && clusterflavor != nil {
			if reply.Action == edgeproto.NoticeAction_UPDATE {
				log.DebugLog(log.DebugLevelNotify,
					"client cluster flavor update",
					"key", clusterflavor.Key.GetKeyString())
				recvClusterFlavor.Update(clusterflavor, 0)
			} else if reply.Action == edgeproto.NoticeAction_DELETE {
				log.DebugLog(log.DebugLevelNotify,
					"client cluster flavor delete",
					"key", clusterflavor.Key.GetKeyString())
				recvClusterFlavor.Delete(clusterflavor, 0)
			}
			s.stats.ClusterFlavorRecv++
			s.stats.Recv++
		}
		clusterInst := reply.GetClusterInst()
		if recvClusterInst != nil && clusterInst != nil {
			if reply.Action == edgeproto.NoticeAction_UPDATE {
				log.DebugLog(log.DebugLevelNotify,
					"client cluster inst update",
					"key", clusterInst.Key.GetKeyString())
				recvClusterInst.Update(clusterInst, 0)
			} else if reply.Action == edgeproto.NoticeAction_DELETE {
				log.DebugLog(log.DebugLevelNotify,
					"client cluster inst delete",
					"key", clusterInst.Key.GetKeyString())
				recvClusterInst.Delete(clusterInst, 0)
			}
			s.stats.ClusterInstRecv++
			s.stats.Recv++
		}
		appInst := reply.GetAppInst()
		if recvAppInst != nil && appInst != nil {
			if reply.Action == edgeproto.NoticeAction_UPDATE {
				log.DebugLog(log.DebugLevelNotify,
					"client app inst update",
					"key", appInst.Key.GetKeyString())
				recvAppInst.Update(appInst, 0)
			} else if reply.Action == edgeproto.NoticeAction_DELETE {
				log.DebugLog(log.DebugLevelNotify,
					"client app inst delete",
					"key", appInst.Key.GetKeyString())
				recvAppInst.Delete(appInst, 0)
			}
			s.stats.AppInstRecv++
			s.stats.Recv++
		}
		cloudlet := reply.GetCloudlet()
		if recvCloudlet != nil && cloudlet != nil {
			if reply.Action == edgeproto.NoticeAction_UPDATE {
				log.DebugLog(log.DebugLevelNotify,
					"client cloudlet update",
					"key", cloudlet.Key.GetKeyString())
				recvCloudlet.Update(cloudlet, 0)
			} else if reply.Action == edgeproto.NoticeAction_DELETE {
				log.DebugLog(log.DebugLevelNotify,
					"client cloudlet delete",
					"key", cloudlet.Key.GetKeyString())
				recvCloudlet.Delete(cloudlet, 0)
			}
			s.stats.CloudletRecv++
			s.stats.Recv++
		}
	}
	close(s.recvRunning)
}

func (s *Client) send(stream edgeproto.NotifyApi_StreamNoticeClient) {
	var notice edgeproto.NoticeRequest
	var nAppInstInfo edgeproto.NoticeRequest_AppInstInfo
	var nClusterInstInfo edgeproto.NoticeRequest_ClusterInstInfo
	var nCloudletInfo edgeproto.NoticeRequest_CloudletInfo
	var nMetric edgeproto.NoticeRequest_Metric
	var nNode edgeproto.NoticeRequest_Node
	var appInstInfo edgeproto.AppInstInfo
	var clusterInstInfo edgeproto.ClusterInstInfo
	var cloudletInfo edgeproto.CloudletInfo
	var node edgeproto.Node
	var err error

	sendAll := true
	nAppInstInfo.AppInstInfo = &appInstInfo
	nClusterInstInfo.ClusterInstInfo = &clusterInstInfo
	nCloudletInfo.CloudletInfo = &cloudletInfo
	nNode.Node = &node
	sendAppInstInfo := s.handler.SendAppInstInfoHandler()
	sendClusterInstInfo := s.handler.SendClusterInstInfoHandler()
	sendCloudletInfo := s.handler.SendCloudletInfoHandler()
	sendNode := s.handler.SendNodeHandler()
	// trigger initial sendAll
	s.wakeupSend()

	for !s.done && err == nil {
		select {
		case <-s.signal:
		case <-stream.Context().Done():
		}
		s.mux.Lock()
		if len(s.appInstInfos) == 0 && len(s.cloudletInfos) == 0 &&
			len(s.clusterInstInfos) == 0 && len(s.metrics) == 0 &&
			len(s.nodes) == 0 &&
			!s.done && !sendAll && stream.Context().Err() == nil {
			s.mux.Unlock()
			continue
		}
		appInstInfos := s.appInstInfos
		clusterInstInfos := s.clusterInstInfos
		cloudletInfos := s.cloudletInfos
		metrics := s.metrics
		nodes := s.nodes
		s.appInstInfos = make(map[edgeproto.AppInstKey]struct{})
		s.clusterInstInfos = make(map[edgeproto.ClusterInstKey]struct{})
		s.cloudletInfos = make(map[edgeproto.CloudletKey]struct{})
		s.nodes = make(map[edgeproto.NodeKey]struct{})
		s.metrics = make([]*edgeproto.Metric, 0)
		s.mux.Unlock()
		if s.done || stream.Context().Err() != nil {
			break
		}
		if sendAll {
			log.DebugLog(log.DebugLevelNotify, "Notify client send all",
				"server", s.GetServerAddr(),
				"local", s.GetLocalAddr())
			if sendAppInstInfo != nil {
				sendAppInstInfo.GetAllKeys(appInstInfos)
			}
			if sendCloudletInfo != nil {
				sendCloudletInfo.GetAllKeys(cloudletInfos)
			}
			if sendClusterInstInfo != nil {
				sendClusterInstInfo.GetAllKeys(clusterInstInfos)
			}
			if sendNode != nil {
				sendNode.GetAllKeys(nodes)
			}
		}

		if sendNode != nil {
			notice.Data = &nNode
			for key, _ := range nodes {
				found := sendNode.Get(&key, &node)
				if found {
					notice.Action = edgeproto.NoticeAction_UPDATE
				} else {
					notice.Action = edgeproto.NoticeAction_DELETE
					node.Key = key
				}
				log.DebugLog(log.DebugLevelNotify, "Send Node",
					"server", s.GetServerAddr(),
					"action", notice.Action,
					"key", node.Key.GetKeyString())
				err = stream.Send(&notice)
				if err != nil {
					break
				}
				s.stats.NodesSent++
			}
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		if sendClusterInstInfo != nil {
			notice.Data = &nClusterInstInfo
			for key, _ := range clusterInstInfos {
				found := sendClusterInstInfo.Get(&key, &clusterInstInfo)
				if found {
					notice.Action = edgeproto.NoticeAction_UPDATE
				} else {
					notice.Action = edgeproto.NoticeAction_DELETE
					clusterInstInfo.Key = key
				}
				log.DebugLog(log.DebugLevelNotify, "Send ClusterInstInfo",
					"server", s.GetServerAddr(),
					"action", notice.Action,
					"key", clusterInstInfo.Key.GetKeyString())
				err = stream.Send(&notice)
				if err != nil {
					break
				}
				s.stats.ClusterInstInfosSent++
			}
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		if sendAppInstInfo != nil {
			notice.Data = &nAppInstInfo
			for key, _ := range appInstInfos {
				found := sendAppInstInfo.Get(&key, &appInstInfo)
				if found {
					notice.Action = edgeproto.NoticeAction_UPDATE
				} else {
					notice.Action = edgeproto.NoticeAction_DELETE
					appInstInfo.Key = key
				}
				log.DebugLog(log.DebugLevelNotify, "Send AppInstInfo",
					"server", s.GetServerAddr(),
					"action", notice.Action,
					"key", appInstInfo.Key.GetKeyString())
				err = stream.Send(&notice)
				if err != nil {
					break
				}
				s.stats.AppInstInfosSent++
			}
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		if sendCloudletInfo != nil {
			notice.Data = &nCloudletInfo
			for key, _ := range cloudletInfos {
				found := sendCloudletInfo.Get(&key, &cloudletInfo)
				if found {
					notice.Action = edgeproto.NoticeAction_UPDATE
				} else {
					notice.Action = edgeproto.NoticeAction_DELETE
					cloudletInfo.Key = key
				}
				log.DebugLog(log.DebugLevelNotify, "Send CloudletInfo",
					"server", s.GetServerAddr(),
					"action", notice.Action,
					"key", cloudletInfo.Key.GetKeyString())
				err = stream.Send(&notice)
				if err != nil {
					break
				}
				s.stats.CloudletInfosSent++
			}
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		for ii, _ := range metrics {
			nMetric.Metric = metrics[ii]
			notice.Action = edgeproto.NoticeAction_UPDATE
			notice.Data = &nMetric
			err = stream.Send(&notice)
			if err != nil {
				break
			}
			s.stats.MetricsSent++
		}
		if err != nil {
			s.stats.SendErrors++
			break
		}

		if sendAll {
			sendAll = false
		}
		appInstInfos = nil
		clusterInstInfos = nil
		cloudletInfos = nil
		metrics = nil
		nodes = nil
	}
	close(s.sendRunning)
}

func (s *Client) logDisconnect(err error) {
	st, ok := status.FromError(err)
	if err == context.Canceled || (ok && st.Code() == codes.Canceled || err == nil) {
		log.DebugLog(log.DebugLevelNotify, "Notify client connection closed",
			"server", s.GetServerAddr(), "local", s.GetLocalAddr(),
			"err", err)
	} else {
		log.InfoLog("Notify client connection failed",
			"server", s.GetServerAddr(), "local", s.GetLocalAddr(),
			"err", err)
	}
}

func (s *Client) UpdateAppInstInfo(key *edgeproto.AppInstKey, old *edgeproto.AppInstInfo) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.appInstInfos[*key] = struct{}{}
	s.wakeupSend()
}

func (s *Client) UpdateClusterInstInfo(key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInstInfo) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.clusterInstInfos[*key] = struct{}{}
	s.wakeupSend()
}

func (s *Client) UpdateCloudletInfo(key *edgeproto.CloudletKey, old *edgeproto.CloudletInfo) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.cloudletInfos[*key] = struct{}{}
	s.wakeupSend()
}

func (s *Client) UpdateNode(key *edgeproto.NodeKey, old *edgeproto.Node) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.nodes[*key] = struct{}{}
	s.wakeupSend()
}

func (s *Client) SendMetric(metric *edgeproto.Metric) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if len(s.metrics) > MetricsLimit {
		s.stats.MetricsDropped++
		return
	}
	s.metrics = append(s.metrics, metric)
	s.wakeupSend()
}

func (s *Client) wakeupSend() {
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

func (s *Client) GetStats(stats *ClientStats) {
	*stats = s.stats
}

func (s *Client) GetServerAddr() string {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.addrs[s.addrIdx]
}

func (s *Client) SetLocalAddr(addr string) {
	s.localAddrMux.Lock()
	defer s.localAddrMux.Unlock()
	s.localAddr = addr
}

func (s *Client) GetLocalAddr() string {
	s.localAddrMux.Lock()
	defer s.localAddrMux.Unlock()
	return s.localAddr
}

// grpcStatsHandler is a grpc interceptor that allows us to get the local
// address of the grpc connection.
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
