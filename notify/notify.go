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
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var NotifyRetryTime time.Duration = 250 * time.Millisecond

const NotifyVersion uint32 = 1

type NotifyType int

const (
	NotifyTypeMatcher NotifyType = iota
	NotifyTypeCloudletMgr
)

type NotifySendHandler interface {
	// Get all the keys for known app insts.
	// The value associated with the key is ignored.
	GetAllAppInstKeys(keys map[proto.AppInstKey]bool)
	// Copy back the value for the app inst.
	// If the app inst was not found, return false instead of true.
	GetAppInst(key *proto.AppInstKey, buf *proto.AppInst) bool
	// Get all the keys for known cloudlets.
	// The value associated with the key is ignored.
	GetAllCloudletKeys(keys map[proto.CloudletKey]bool)
	// Copy back the value for the cloudlet.
	// If the cloudlet was not found, return false instead of true.
	GetCloudlet(key *proto.CloudletKey, buf *proto.Cloudlet) bool
}

type NotifySenderStats struct {
	AppInstsSent  uint64
	CloudletsSent uint64
	Connects      uint64
}

type NotifySender struct {
	addr      string
	appInsts  map[proto.AppInstKey]bool
	cloudlets map[proto.CloudletKey]bool
	mux       util.Mutex
	cond      sync.Cond
	done      bool
	handler   NotifySendHandler
	stats     NotifySenderStats
	version   uint32
	ntype     NotifyType
	running   chan int
}

type NotifySenders struct {
	table   map[string]*NotifySender
	mux     util.Mutex
	handler NotifySendHandler
}

var notifySenders NotifySenders

func InitNotifySenders(handler NotifySendHandler) {
	notifySenders.table = make(map[string]*NotifySender)
	notifySenders.handler = handler
}

func RegisterMatcherAddrs(addrs string) {
	if addrs == "" {
		return
	}
	list := strings.Split(addrs, ",")
	for _, addr := range list {
		RegisterReceiver(addr, NotifyTypeMatcher)
	}
}

func RegisterCloudletAddrs(addrs string) {
	if addrs == "" {
		return
	}
	list := strings.Split(addrs, ",")
	for _, addr := range list {
		RegisterReceiver(addr, NotifyTypeCloudletMgr)
	}
}

func RegisterReceiver(addr string, ntype NotifyType) {
	notifySenders.mux.Lock()
	defer notifySenders.mux.Unlock()

	_, found := notifySenders.table[addr]
	if found {
		return
	}
	notifier := &NotifySender{}
	notifier.addr = addr
	notifier.appInsts = make(map[proto.AppInstKey]bool)
	notifier.mux.InitCond(&notifier.cond)
	notifier.handler = notifySenders.handler
	notifier.ntype = ntype
	notifier.running = make(chan int)
	notifySenders.table[addr] = notifier
	go notifier.Run()
}

func UnregisterReceiver(addr string) {
	notifySenders.mux.Lock()
	notifier, found := notifySenders.table[addr]
	if found {
		delete(notifySenders.table, addr)
	}
	notifySenders.mux.Unlock()
	if notifier == nil {
		return
	}
	notifier.Stop()
}

func GetNotifySenderStats(addr string) *NotifySenderStats {
	stats := &NotifySenderStats{}
	notifySenders.mux.Lock()
	defer notifySenders.mux.Unlock()
	notifier, found := notifySenders.table[addr]
	if found {
		stats = &NotifySenderStats{}
		*stats = notifier.stats
	}
	return stats
}

func UpdateAppInst(key *proto.AppInstKey) {
	notifySenders.mux.Lock()
	defer notifySenders.mux.Unlock()
	for _, notifier := range notifySenders.table {
		if notifier.ntype != NotifyTypeMatcher {
			continue
		}
		notifier.UpdateAppInst(key)
	}
}

func UpdateCloudlet(key *proto.CloudletKey) {
	notifySenders.mux.Lock()
	defer notifySenders.mux.Unlock()
	for _, notifier := range notifySenders.table {
		if notifier.ntype != NotifyTypeCloudletMgr {
			continue
		}
		notifier.UpdateCloudlet(key)
	}
}

func (s *NotifySender) UpdateAppInst(key *proto.AppInstKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.appInsts[*key] = true
	s.cond.Signal()
}

func (s *NotifySender) UpdateCloudlet(key *proto.CloudletKey) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.cloudlets[*key] = true
	s.cond.Signal()
}

func (s *NotifySender) Run() {
	var conn *grpc.ClientConn
	var sendAll bool
	var notice proto.Notice
	var noticeAppInst proto.Notice_AppInst
	var noticeCloudlet proto.Notice_Cloudlet
	var appInst proto.AppInst
	var cloudlet proto.Cloudlet
	var client proto.NotifyApi_StreamNoticeClient
	var reply *proto.NoticeReply
	var err error
	var sendErr error

	noticeAppInst.AppInst = &appInst
	noticeCloudlet.Cloudlet = &cloudlet
	tries := 0
	for !s.done {
		if sendErr != nil {
			util.DebugLog(util.DebugLevelNotify, "NotifySender sendErr", "addr", s.addr, "error", err)
			conn.Close()
			client = nil
			sendErr = nil
		}
		if client == nil {
			// connect to receiver
			tries++
			ctx, cancel := context.WithTimeout(context.Background(), NotifyRetryTime)
			conn, err = grpc.DialContext(ctx, s.addr, grpc.WithInsecure())
			cancel()
			if err != nil {
				if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
					time.Sleep(NotifyRetryTime)
				}
				continue
			}
			api := proto.NewNotifyApiClient(conn)
			client, err = api.StreamNotice(context.Background())
			if err != nil {
				util.DebugLog(util.DebugLevelNotify, "NotifySender get client", "addr", s.addr, "error", err)
				conn.Close()
				client = nil
				time.Sleep(NotifyRetryTime)
				continue
			}
			s.stats.Connects++

			// Send our version and read back remote version.
			// We use the lowest common version.
			notice.Version = NotifyVersion
			notice.Action = proto.NoticeAction_VERSION
			notice.ConnectionId = s.stats.Connects
			notice.Data = nil
			sendErr = client.Send(&notice)
			if sendErr != nil {
				continue
			}
			reply, sendErr = client.Recv()
			if sendErr != nil {
				continue
			}
			if notice.Version > reply.Version {
				s.version = reply.Version
			} else {
				s.version = notice.Version
			}
			util.DebugLog(util.DebugLevelNotify, "NotifySender connected", "addr", s.addr, "version", s.version, "supported-version", NotifyVersion, "tries", tries)
			tries = 0
			sendAll = true
		}
		s.mux.Lock()
		for len(s.appInsts) == 0 && len(s.cloudlets) == 0 && !s.done && !sendAll {
			s.cond.Wait()
		}
		appInsts := s.appInsts
		cloudlets := s.cloudlets
		s.appInsts = make(map[proto.AppInstKey]bool)
		s.cloudlets = make(map[proto.CloudletKey]bool)
		s.mux.Unlock()
		if s.done {
			break
		}
		if sendAll {
			util.DebugLog(util.DebugLevelNotify, "Send all", "addr", s.addr)
			if s.ntype == NotifyTypeMatcher {
				s.handler.GetAllAppInstKeys(appInsts)
			}
			if s.ntype == NotifyTypeCloudletMgr {
				s.handler.GetAllCloudletKeys(cloudlets)
			}
		}
		// send appInsts
		notice.Data = &noticeAppInst
		for key, _ := range appInsts {
			found := s.handler.GetAppInst(&key, &appInst)
			if found {
				notice.Action = proto.NoticeAction_UPDATE
			} else {
				notice.Action = proto.NoticeAction_DELETE
				appInst.Key = key
			}
			util.DebugLog(util.DebugLevelNotify, "Send app inst", "addr", s.addr, "action", notice.Action, "key", appInst.Key.GetKeyString())
			sendErr = client.Send(&notice)
			if sendErr != nil {
				break
			}
			s.stats.AppInstsSent++
		}
		// send cloudlets
		notice.Data = &noticeCloudlet
		for key, _ := range cloudlets {
			found := s.handler.GetCloudlet(&key, &cloudlet)
			if found {
				notice.Action = proto.NoticeAction_UPDATE
			} else {
				notice.Action = proto.NoticeAction_DELETE
				cloudlet.Key = key
			}
			util.DebugLog(util.DebugLevelNotify, "Send cloudlet", "addr", s.addr, "action", notice.Action, "key", cloudlet.Key.GetKeyString())
			sendErr = client.Send(&notice)
			if sendErr != nil {
				break
			}
			s.stats.CloudletsSent++
		}
		if sendAll && sendErr == nil {
			notice.Action = proto.NoticeAction_SENDALL_END
			notice.Data = nil
			sendErr = client.Send(&notice)
			if sendErr != nil {
				break
			}
			sendAll = false
		}
		appInsts = nil
		cloudlets = nil
		if sendErr != nil {
			continue
		}
	}
	if client != nil {
		conn.Close()
	}
	close(s.running)
}

func (s *NotifySender) Stop() {
	s.mux.Lock()
	s.done = true
	s.cond.Signal()
	s.mux.Unlock()
	<-s.running
}

type NotifySendAllMaps struct {
	appInsts  map[proto.AppInstKey]bool
	cloudlets map[proto.CloudletKey]bool
}

type NotifyRecvHandler interface {
	// NotifySendAllMaps contains all the keys that were updated in
	// the initial send of all data. Any keys that were not sent should
	// be purged from local memory since they no longer exist on the
	// sender.
	HandleSendAllDone(allMaps *NotifySendAllMaps)
	// Handle an update or delete notice from the sender
	HandleNotice(notice *proto.Notice) error
}

type NotifyReceiver struct {
	network      string
	address      string
	handler      NotifyRecvHandler
	server       *grpc.Server
	version      uint32
	connectionId uint64
	mux          util.Mutex
	done         bool
	running      chan int
}

func NewNotifyReceiver(network, address string, handler NotifyRecvHandler) *NotifyReceiver {
	s := &NotifyReceiver{}
	s.network = network
	s.address = address
	s.handler = handler
	return s
}

func (s *NotifyReceiver) Run() {
	s.mux.Lock()
	if s.running != nil || s.done {
		// already another thread running this or we're already done
		s.mux.Unlock()
		return
	}
	s.running = make(chan int)
	s.mux.Unlock()

	for !s.done {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			util.InfoLog("Failed to listen", "address", s.network+":"+s.address, "error", err)
			time.Sleep(NotifyRetryTime)
			continue
		}
		// to avoid race conditions with NotifyReceiver.Stop() calling
		// s.server.Stop() when s.server is being modified, we always
		// hold the lock when modifying s.server.
		s.mux.Lock()
		if s.done {
			lis.Close()
			s.mux.Unlock()
			break
		}
		s.server = grpc.NewServer()
		s.mux.Unlock()
		proto.RegisterNotifyApiServer(s.server, s)

		err = s.server.Serve(lis)
		util.DebugLog(util.DebugLevelNotify, "NotifyReceiver serve", "error", err)
		if !s.done {
			time.Sleep(NotifyRetryTime)
		}
	}
	close(s.running)
}

func (s *NotifyReceiver) Stop() {
	s.mux.Lock()
	s.done = true
	if s.server != nil {
		s.server.Stop()
	}
	s.mux.Unlock()
	// now that we've set done, s.running will never change
	if s.running != nil {
		<-s.running
	}
	s.done = false
	s.running = nil
}

func (s *NotifyReceiver) StreamNotice(stream proto.NotifyApi_StreamNoticeServer) error {
	var notice *proto.Notice
	var reply proto.NoticeReply
	var err error

	defer func() {
		// handle failure
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.Canceled {
				util.DebugLog(util.DebugLevelNotify, "NotifyReceiver stream cancelled", "error", err)
			} else {
				util.InfoLog("NotifyReceiver stream", "error", err)
			}
		}
	}()

	// initial connection is version exchange
	// this also sets the connection Id so we can ignore spurious old
	// buffered messages
	notice, err = stream.Recv()
	if err != nil {
		return err
	}
	if notice.Action != proto.NoticeAction_VERSION {
		util.DebugLog(util.DebugLevelNotify, "NotifyReceiver bad action", "expected", proto.NoticeAction_VERSION, "got", notice.Action)
		return errors.New("NotifyReceiver expected action version")
	}
	// use lowest common version
	if notice.Version > NotifyVersion {
		s.version = notice.Version
	} else {
		s.version = NotifyVersion
	}
	s.connectionId = notice.ConnectionId
	// send back my version
	reply.Action = proto.NoticeAction_VERSION
	reply.Version = s.version
	err = stream.Send(&reply)
	if err != nil {
		return err
	}
	sendAllMaps := &NotifySendAllMaps{}
	sendAllMaps.appInsts = make(map[proto.AppInstKey]bool)
	sendAllMaps.cloudlets = make(map[proto.CloudletKey]bool)
	util.DebugLog(util.DebugLevelNotify, "NotifyReceiver connected", "version", s.version, "supported-version", NotifyVersion)

	for !s.done {
		notice, err = stream.Recv()
		if s.done {
			break
		}
		if err != nil {
			return err
		}
		if notice.ConnectionId != s.connectionId {
			return errors.New("Bad connection id")
		}
		if sendAllMaps != nil && notice.Action == proto.NoticeAction_UPDATE {
			appInst := notice.GetAppInst()
			if appInst != nil {
				sendAllMaps.appInsts[appInst.Key] = true
			}
			cloudlet := notice.GetCloudlet()
			if cloudlet != nil {
				sendAllMaps.cloudlets[cloudlet.Key] = true
			}
		}
		if notice.Action == proto.NoticeAction_SENDALL_END {
			s.handler.HandleSendAllDone(sendAllMaps)
			sendAllMaps = nil
			continue
		}
		if notice.Action != proto.NoticeAction_UPDATE && notice.Action != proto.NoticeAction_DELETE {
			return errors.New("Unexpected notice action, not update or delete")
		}
		err = s.handler.HandleNotice(notice)
		if err != nil {
			return err
		}
	}
	return nil
}
