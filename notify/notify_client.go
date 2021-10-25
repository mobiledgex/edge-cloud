package notify

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type Client struct {
	sendrecv      SendRecv
	addrs         []string
	tlsDialOption grpc.DialOption
	version       uint32
	stats         ClientStats
	addrIdx       int
	mux           util.Mutex
	conn          *grpc.ClientConn
	cancel        context.CancelFunc
	running       chan struct{}
	// localAddr has its own lock because its called in some
	// grpc interceptor context
	localAddr    string
	localAddrMux util.Mutex
	name         string
	options      ClientOptions
}

type ClientStats struct {
}

func cancelNoop() {}

func NewClient(name string, addrs []string, tlsDialOption grpc.DialOption, ops ...ClientOp) *Client {
	s := Client{}
	s.name = name
	s.addrs = addrs
	if tlsDialOption == nil {
		tlsDialOption = grpc.WithInsecure()
	}
	s.tlsDialOption = tlsDialOption
	for _, op := range ops {
		op(&s.options)
	}
	s.sendrecv.init("client")
	return &s
}

func (s *Client) SetFilterByCloudletKey() {
	s.sendrecv.filterCloudletKeys = true
}

func (s *Client) Start() {
	s.mux.Lock()
	s.sendrecv.done = false
	s.cancel = cancelNoop
	s.running = make(chan struct{})
	s.addrIdx = rand.Int() % len(s.addrs)
	s.mux.Unlock()
	go s.run()
}

func (s *Client) Stop() {
	s.mux.Lock()
	s.sendrecv.done = true
	s.cancel()
	s.mux.Unlock()
	<-s.running
}

func (s *Client) RegisterSend(send NotifySend) {
	s.sendrecv.registerSend(send)
}

func (s *Client) RegisterRecv(recv NotifyRecv) {
	s.sendrecv.registerRecv(recv)
}

func (s *Client) RegisterSendAllRecv(handler SendAllRecv) {
	s.sendrecv.registerSendAllRecv(handler)
}

func (s *Client) run() {
	for !s.sendrecv.done {
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

		s.sendrecv.sendRunning = make(chan struct{})
		s.sendrecv.recvRunning = make(chan struct{})
		go s.sendrecv.send(stream)
		go s.sendrecv.recv(stream, 0, CleanupPrune)
		// if there is a communication error, both threads will exit
		<-s.sendrecv.sendRunning
		<-s.sendrecv.recvRunning
		s.logDisconnect(stream.Context().Err())
		s.connectCleanup()
	}
	log.DebugLog(log.DebugLevelNotify, "Notify client stopped",
		"server", s.GetServerAddr(), "local", s.GetLocalAddr())
	close(s.running)
}

// connect to the server
func (s *Client) connect() (StreamNotify, error) {
	// connect to server
	s.mux.Lock()
	s.sendrecv.stats.Tries++
	s.addrIdx++
	if s.addrIdx == len(s.addrs) {
		s.addrIdx = 0
	}
	addr := s.addrs[s.addrIdx]
	s.mux.Unlock()

	span := log.StartSpan(log.DebugLevelNotify, "notify client connect")
	defer span.Finish()
	spanctx := log.ContextWithSpan(context.Background(), span)

	ctx, cancel := context.WithTimeout(spanctx, NotifyRetryTime)
	conn, err := grpc.DialContext(ctx, addr,
		s.tlsDialOption,
		grpc.WithStatsHandler(&grpcStatsHandler{client: s}),
		grpc.WithKeepaliveParams(clientParams),
		grpc.WithChainUnaryInterceptor(s.options.unaryInterceptors...),
		grpc.WithChainStreamInterceptor(s.options.streamInterceptors...),
	)
	cancel()
	if err != nil {
		return nil, err
	}

	api := edgeproto.NewNotifyApiClient(conn)
	ctx, cancel = context.WithCancel(spanctx)
	stream, err := api.StreamNotice(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "Notify client get stream",
			"server", addr, "error", err)
		cancel()
		conn.Close()
		return nil, err
	}
	s.mux.Lock()
	s.cancel = cancel
	s.conn = conn
	s.sendrecv.stats.Connects++
	s.mux.Unlock()

	ctx = stream.Context()
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("Notify client unable to get peer context")
	}
	s.sendrecv.peerAddr = peer.Addr.String()

	return stream, nil
}

func (s *Client) connectCleanup() {
	s.mux.Lock()
	s.conn.Close()
	s.sendrecv.peer = ""
	s.cancel()
	s.cancel = cancelNoop
	s.mux.Unlock()
	s.SetLocalAddr("")
}

// negotiate performs the initial exchange between client and server
// that register what type of client we are with the server, and
// settle upon the maximum supported version.
func (s *Client) negotiate(stream StreamNotify) error {
	var request edgeproto.Notice
	var reply *edgeproto.Notice
	// Send our version and read back remote version.
	// We use the lowest common version.
	request.Version = NotifyVersion
	request.Action = edgeproto.NoticeAction_VERSION
	request.WantObjs = s.sendrecv.localWanted
	request.FilterCloudletKey = s.sendrecv.filterCloudletKeys
	request.Tags = map[string]string{
		"name": s.name,
	}
	err := stream.Send(&request)
	if err != nil {
		s.sendrecv.stats.NegotiateErrors++
		return err
	}
	reply, err = stream.Recv()
	if err != nil {
		s.sendrecv.stats.NegotiateErrors++
		return err
	}
	if request.Version > reply.Version {
		s.version = reply.Version
	} else {
		s.version = request.Version
	}
	s.sendrecv.setRemoteWanted(reply.WantObjs)
	if reply.Tags != nil {
		if peer, found := reply.Tags["name"]; found {
			s.sendrecv.peer = peer
		}
	}

	s.mux.Lock()
	addr := s.addrs[s.addrIdx]
	s.mux.Unlock()
	log.DebugLog(log.DebugLevelNotify, "Notify client connected",
		"server", addr, "peer", s.sendrecv.peer, "local", s.GetLocalAddr(),
		"version", s.version, "supported-version", NotifyVersion,
		"remoteWanted", s.sendrecv.remoteWanted,
		"filterCloudletKey", s.sendrecv.filterCloudletKeys,
		"tries", s.sendrecv.stats.Tries,
		"connects", s.sendrecv.stats.Connects)
	return nil
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

func (s *Client) GetStats(stats *Stats) {
	*stats = s.sendrecv.stats
	s.sendrecv.setObjStats(stats)
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

type ClientOptions struct {
	unaryInterceptors  []grpc.UnaryClientInterceptor
	streamInterceptors []grpc.StreamClientInterceptor
}

type ClientOp func(s *ClientOptions)

func ClientUnaryInterceptors(i ...grpc.UnaryClientInterceptor) ClientOp {
	return func(opts *ClientOptions) {
		opts.unaryInterceptors = append(opts.unaryInterceptors, i...)
	}
}

func ClientStreamInterceptors(i ...grpc.StreamClientInterceptor) ClientOp {
	return func(opts *ClientOptions) {
		opts.streamInterceptors = append(opts.streamInterceptors, i...)
	}
}

type ClientState struct {
	Name              string
	Peer              string
	NegotiatedVersion uint32
	Stats             Stats
}

func (s *Client) GetState() *ClientState {
	s.mux.Lock()
	defer s.mux.Unlock()
	state := ClientState{}
	state.Name = s.name
	if s.sendrecv.peer != "" {
		state.Peer = s.sendrecv.peer
	} else {
		state.Peer = "not connected"
	}
	state.NegotiatedVersion = s.version
	state.Stats = s.sendrecv.stats
	return &state
}
