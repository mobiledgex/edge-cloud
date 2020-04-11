package crmutil

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util/proxyutil"
	"github.com/mobiledgex/edge-cloud/util/webrtcutil"
	ssh "github.com/mobiledgex/golang-ssh"
	opentracing "github.com/opentracing/opentracing-go"
	webrtc "github.com/pion/webrtc/v2"
	"github.com/xtaci/smux"
)

// ExecReqHandler just satisfies the Recv() function for the
// ExecRequest receive notify interface, and calls into the
// controller data which has all the cached information about the
// ClusterInst, AppInst, etc.
type ExecReqHandler struct {
	cd *ControllerData
}

func NewExecReqHandler(cd *ControllerData) *ExecReqHandler {
	return &ExecReqHandler{cd: cd}
}

func (s *ExecReqHandler) RecvExecRequest(ctx context.Context, msg *edgeproto.ExecRequest) {
	// spawn go process so we don't stall notify messages
	go func() {
		cspan := log.StartSpan(log.DebugLevelApi, "process exec req", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
		defer cspan.Finish()
		var err error
		if msg.VmType == edgeproto.VMType_UNKNOWN_VM {
			err = s.cd.ProcessExecReq(ctx, msg)
		} else {
			err = s.cd.ProcessAccessCloudlet(ctx, msg)
		}
		if err != nil {
			msg.Err = err.Error()
		}
		s.cd.ExecReqSend.Update(ctx, msg)
	}()
}

func (cd *ControllerData) ProcessExecReq(ctx context.Context, req *edgeproto.ExecRequest) error {
	var err error

	run := &WebrtcExec{
		req: req,
	}

	appInst := edgeproto.AppInst{}
	app := edgeproto.App{}
	found := cd.AppInstCache.Get(&req.AppInstKey, &appInst)
	if !found {
		return fmt.Errorf("app inst %s not found",
			req.AppInstKey.GetKeyString())
	}
	found = cd.AppCache.Get(&req.AppInstKey.AppKey, &app)
	if !found {
		return fmt.Errorf("app %s not found",
			req.AppInstKey.AppKey.GetKeyString())
	}

	execReqInfo := cloudcommon.ExecReqInfo{}
	if req.Console != nil {
		req.Console.Url, err = cd.platform.GetConsoleUrl(ctx, &app)
		if err != nil {
			return err
		}
		urlObj, err := url.Parse(req.Console.Url)
		if err != nil {
			return fmt.Errorf("unable to parse console url, %s, %v", req.Console.Url, err)
		}
		execReqInfo.Type = cloudcommon.ExecReqConsole
		execReqInfo.InitURL = urlObj
	} else {
		clusterInst := edgeproto.ClusterInst{}
		found := cd.ClusterInstCache.Get(&appInst.Key.ClusterInstKey, &clusterInst)
		if !found {
			return fmt.Errorf("cluster inst %s not found",
				appInst.Key.ClusterInstKey.GetKeyString())
		}
		run.client, err = cd.platform.GetPlatformClientRootLB(ctx, &clusterInst)
		if err != nil {
			return err
		}

		run.contcmd, err = cd.platform.GetContainerCommand(ctx, &clusterInst, &app, &appInst, req)
		if err != nil {
			return err
		}
		execReqInfo.Type = cloudcommon.ExecReqConsole
	}

	if !req.Webrtc {
		return cd.SetupEdgeTurnConn(ctx, req, &execReqInfo, run)
	}

	offer := webrtc.SessionDescription{}
	err = json.Unmarshal([]byte(req.Offer), &offer)
	if err != nil {
		return fmt.Errorf("unable to decode offer, %v", err)
	}

	// hard code config for now
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.mobiledgex.net:19302"},
			},
		},
	}

	// create a new peer connection
	peerConn, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi,
			"failed to establish peer connection",
			"config", config, "err", err)
		return fmt.Errorf("failed to establish peer connection, %v", err)
	}

	// register handlers
	if req.Console != nil {
		peerConn.OnDataChannel(run.RTCTunnel)
	} else {
		peerConn.OnDataChannel(run.DataChannel)
	}

	// set remote description
	err = peerConn.SetRemoteDescription(offer)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi,
			"failed to set remote description",
			"offer", offer, "peerConn", peerConn, "err", err)
		return fmt.Errorf("failed to set remote description, %v", err)
	}
	// create answer
	answer, err := peerConn.CreateAnswer(nil)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi,
			"failed to create answer",
			"peerConn", peerConn, "err", err)
		return fmt.Errorf("failed to set answer, %v", err)
	}
	// set local description, and starts out UDP listeners
	err = peerConn.SetLocalDescription(answer)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi,
			"failed to set local description",
			"peerConn", peerConn, "err", err)
		return fmt.Errorf("failed to set local description, %v", err)
	}

	// send back answer
	answerBytes, err := json.Marshal(answer)
	if err != nil {
		return fmt.Errorf("failed to encode answer, %v", err)
	}
	req.Answer = string(answerBytes)
	log.SpanLog(ctx, log.DebugLevelApi, "returning answer")

	return nil
}

func (cd *ControllerData) ProcessAccessCloudlet(ctx context.Context, req *edgeproto.ExecRequest) error {
	var err error

	run := &WebrtcExec{
		req: req,
	}

	execReqInfo := cloudcommon.ExecReqInfo{
		Type: cloudcommon.ExecReqShell,
	}

	switch req.VmType {
	case edgeproto.VMType_DEDICATED_ROOTLB_VM:
		clusterInst := edgeproto.ClusterInst{}
		clusterInstKey := &req.AppInstKey.ClusterInstKey
		found := cd.ClusterInstCache.Get(clusterInstKey, &clusterInst)
		if !found {
			return fmt.Errorf("cluster inst %s not found",
				clusterInstKey.GetKeyString())
		}
		if clusterInst.IpAccess != edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			return fmt.Errorf("cluster inst %s does not have ipaccess set to dedicated", clusterInstKey.GetKeyString())
		}
		run.client, err = cd.platform.GetPlatformClientRootLB(ctx, &clusterInst)
		if err != nil {
			return err
		}
	case edgeproto.VMType_SHARED_ROOTLB_VM:
		rlbName := cloudcommon.GetRootLBName(&req.AppInstKey.ClusterInstKey.CloudletKey)
		run.client, err = cd.platform.GetPlatformClient(ctx, rlbName)
		if err != nil {
			return err
		}
	case edgeproto.VMType_PLATFORM_VM:
		pfName := cloudcommon.GetPlatformVMName(&req.AppInstKey.ClusterInstKey.CloudletKey)
		run.client, err = cd.platform.GetPlatformClient(ctx, pfName)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid vmtype: %s", req.VmType)
	}

	return cd.SetupEdgeTurnConn(ctx, req, &execReqInfo, run)
}

func (cd *ControllerData) SetupEdgeTurnConn(ctx context.Context, req *edgeproto.ExecRequest, execReqInfo *cloudcommon.ExecReqInfo, run *WebrtcExec) error {
	if execReqInfo == nil {
		return fmt.Errorf("execreqinfo is nil")
	}
	// Connect to EdgeTurn server
	if req.EdgeTurnAddr == "" {
		return fmt.Errorf("no edgeturn server address specified")
	}

	tlsConfig, err := cd.NodeMgr.InternalPki.GetClientTlsConfig(ctx,
		cd.NodeMgr.CommonName(),
		node.CertIssuerRegionalCloudlet,
		[]node.MatchCA{node.SameRegionalMatchCA()})
	if err != nil {
		return err
	}
	turnConn, err := tls.Dial("tcp", req.EdgeTurnAddr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to edgeturn server: %v", err)
	}
	defer turnConn.Close()

	out, err := json.Marshal(execReqInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal execReqInfo %v, %v", execReqInfo, err)
	}
	turnConn.Write(out)
	log.SpanLog(ctx, log.DebugLevelApi, "sent execreq info", "info", string(out))

	// Fetch session info from EdgeTurn server
	var sessInfo cloudcommon.SessionInfo
	d := json.NewDecoder(turnConn)
	err = d.Decode(&sessInfo)
	if err != nil {
		return fmt.Errorf("failed to decode session info: %v", err)
	}
	log.SpanLog(ctx, log.DebugLevelApi, "received session info from edgeturn server", "info", sessInfo)

	turnAddrParts := strings.Split(req.EdgeTurnAddr, ":")
	if len(turnAddrParts) != 2 {
		return fmt.Errorf("invalid edgeturn Addr: %s", req.EdgeTurnAddr)
	}

	switch execReqInfo.Type {
	case cloudcommon.ExecReqConsole:
		proxyAddr := "https://" + turnAddrParts[0] + ":" + sessInfo.AccessPort + "/edgeconsole?edgetoken=" + sessInfo.Token
		req.AccessUrl = proxyAddr
		cd.ExecReqSend.Update(ctx, req)

		err = proxyutil.ProxyMuxServer(turnConn, req.Console.Url)
		if err != nil {
			return err
		}
	case cloudcommon.ExecReqShell:
		proxyAddr := "wss://" + turnAddrParts[0] + ":" + sessInfo.AccessPort + "/edgeshell?edgetoken=" + sessInfo.Token
		req.AccessUrl = proxyAddr
		cd.ExecReqSend.Update(ctx, req)
		run.proxyRawConn(turnConn)
	default:
		return fmt.Errorf("invalid execreqinfo type: %v", execReqInfo.Type)
	}

	return nil
}

type WebrtcExec struct {
	req     *edgeproto.ExecRequest
	client  ssh.Client
	contcmd string
	sin     io.WriteCloser
}

func (s *WebrtcExec) DataChannel(d *webrtc.DataChannel) {
	d.OnOpen(func() {
		wr := webrtcutil.NewDataChanWriter(d)
		prd, pwr := io.Pipe()
		s.sin = pwr

		err := s.client.Shell(prd, wr, wr, s.contcmd)
		if err != nil {
			log.DebugLog(log.DebugLevelApi,
				"failed to exec",
				"cmd", s.contcmd, "err", err)
		}
		d.Close()
	})
	d.OnMessage(func(msg webrtc.DataChannelMessage) {
		s.sin.Write(msg.Data)
	})
	d.OnClose(func() {
	})
}

func (s *WebrtcExec) RTCTunnel(d *webrtc.DataChannel) {
	d.OnOpen(func() {
		dcconn, err := webrtcutil.WrapDataChannel(d)
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "failed to wrap webrtc datachannel", "err", err)
			return
		}
		s.proxyConsoleConn(s.req.Console.Url, dcconn)
	})
}

func (s *WebrtcExec) proxyConsoleConn(fromUrl string, toConn net.Conn) {
	urlObj, err := url.Parse(fromUrl)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "failed to parse console url", "url", fromUrl, "err", err)
		return
	}
	sess, err := smux.Server(toConn, nil)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "failed to setup smux server", "err", err)
		return
	}
	defer sess.Close()
	log.DebugLog(log.DebugLevelApi, "Successfully started proxy", "fromUrl", fromUrl)
	for {
		stream, err := sess.AcceptStream()
		if err != nil {
			if err.Error() != io.ErrClosedPipe.Error() {
				log.DebugLog(log.DebugLevelApi, "failed to setup smux acceptstream", "err", err)
			}
			return
		}
		var server net.Conn
		if urlObj.Scheme == "http" {
			server, err = net.Dial("tcp", urlObj.Host)
			if err != nil {
				log.DebugLog(log.DebugLevelApi, "failed to get console", "err", err)
				return
			}
		} else if urlObj.Scheme == "https" {
			server, err = tls.Dial("tcp", urlObj.Host, &tls.Config{
				InsecureSkipVerify: true,
			})
			if err != nil {
				log.DebugLog(log.DebugLevelApi, "failed to get console", "err", err)
				return
			}
		} else {
			log.DebugLog(log.DebugLevelApi, "unsupported scheme", "scheme", urlObj.Scheme)
			return
		}
		go func(server net.Conn, stream *smux.Stream) {
			buf := make([]byte, 1500)
			for {
				n, err := stream.Read(buf)
				if err != nil {
					break
				}
				server.Write(buf[:n])
			}
			stream.Close()
			server.Close()
		}(server, stream)

		go func(server net.Conn, stream *smux.Stream) {
			buf := make([]byte, 1500)
			for {
				n, err := server.Read(buf)
				if err != nil {
					break
				}
				stream.Write(buf[:n])
			}
			stream.Close()
			server.Close()
		}(server, stream)
	}
}

func (s *WebrtcExec) proxyRawConn(turnConn net.Conn) {
	prd, pwr := io.Pipe()
	go io.Copy(pwr, turnConn)
	err := s.client.Shell(prd, turnConn, turnConn, s.contcmd)
	if err != nil {
		log.DebugLog(log.DebugLevelApi,
			"failed to exec",
			"cmd", s.contcmd, "err", err)
	}
}
