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
		err := s.cd.ProcessExecReq(ctx, msg)
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
	if req.Cmd == nil || req.Cmd.CloudletMgmtNode == nil {
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
	}

	var execReqType cloudcommon.ExecReqType
	var initURL *url.URL
	if req.Console != nil {
		req.Console.Url, err = cd.platform.GetConsoleUrl(ctx, &app)
		if err != nil {
			return err
		}
		urlObj, err := url.Parse(req.Console.Url)
		if err != nil {
			return fmt.Errorf("unable to parse console url, %s, %v", req.Console.Url, err)
		}
		execReqType = cloudcommon.ExecReqConsole
		initURL = urlObj
	} else if req.Cmd != nil && req.Cmd.CloudletMgmtNode != nil {
		insts := []edgeproto.ClusterInst{}
		cd.ClusterInstCache.Mux.Lock()
		for _, v := range cd.ClusterInstCache.Objs {
			insts = append(insts, *v.Obj)
		}
		cd.ClusterInstCache.Mux.Unlock()
		nodes, err := cd.platform.ListCloudletMgmtNodes(ctx, insts)
		if err != nil {
			return fmt.Errorf("unable to get list of cloudlet mgmt nodes, %v", err)
		}
		access_node := req.Cmd.CloudletMgmtNode
		found := false
		for _, node := range nodes {
			if access_node.Type == "" && access_node.Name != "" {
				// wildcard on node type, so allow node name
				found = true
				break
			}
			if access_node.Type != node.Type {
				continue
			}
			if access_node.Name == "" {
				access_node.Name = node.Name
				found = true
				break
			} else if access_node.Name == node.Name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unable to find cloudlet mgmt node, list of valid nodes: %v", nodes)
		}
		run.contcmd = "bash"
		if req.Cmd.Command != "" {
			run.contcmd = req.Cmd.Command
		}
		run.client, err = cd.platform.GetNodePlatformClient(ctx, access_node)
		if err != nil {
			return err
		}
		execReqType = cloudcommon.ExecReqShell
	} else {
		execReqType = cloudcommon.ExecReqShell
		clusterInst := edgeproto.ClusterInst{}
		found := cd.ClusterInstCache.Get(&appInst.Key.ClusterInstKey, &clusterInst)
		if !found {
			return fmt.Errorf("cluster inst %s not found",
				appInst.Key.ClusterInstKey.GetKeyString())
		}

		run.contcmd, err = cd.platform.GetContainerCommand(ctx, &clusterInst, &app, &appInst, req)
		if err != nil {
			return err
		}

		run.client, err = cd.platform.GetClusterPlatformClient(ctx, &clusterInst)
		if err != nil {
			return err
		}
	}

	if req.Webrtc {
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
	} else {
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

		// Send ExecReqInfo to EdgeTurn server
		execReqInfo := cloudcommon.ExecReqInfo{
			Type:    execReqType,
			InitURL: initURL,
		}
		out, err := json.Marshal(&execReqInfo)
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

		if req.Console != nil {
			urlObj, err := url.Parse(req.Console.Url)
			if err != nil {
				return fmt.Errorf("failed to parse console url %s, %v", req.Console.Url, err)
			}
			isTLS := false
			if urlObj.Scheme == "http" {
				isTLS = false
			} else if urlObj.Scheme == "https" {
				isTLS = true
			} else {
				return fmt.Errorf("unsupported scheme %s", urlObj.Scheme)
			}
			sess, err := smux.Server(turnConn, nil)
			if err != nil {
				return fmt.Errorf("failed to setup smux server, %v", err)
			}
			// Verify if connection to url is okay
			var server net.Conn
			if isTLS {
				server, err = tls.Dial("tcp", urlObj.Host, &tls.Config{
					InsecureSkipVerify: true,
				})
				if err != nil {
					return fmt.Errorf("failed to get console, %v", err)
				}
			} else {
				server, err = net.Dial("tcp", urlObj.Host)
				if err != nil {
					return fmt.Errorf("failed to get console, %v", err)
				}
			}
			server.Close()
			defer sess.Close()
			// Notify controller that connection is setup
			proxyAddr := "https://" + turnAddrParts[0] + ":" + sessInfo.AccessPort + "/edgeconsole?edgetoken=" + sessInfo.Token
			req.AccessUrl = proxyAddr
			cd.ExecReqSend.Update(ctx, req)
			for {
				stream, err := sess.AcceptStream()
				if err != nil {
					if err.Error() != io.ErrClosedPipe.Error() {
						return fmt.Errorf("failed to setup smux acceptstream, %v", err)
					}
					return nil
				}
				if isTLS {
					server, err = tls.Dial("tcp", urlObj.Host, &tls.Config{
						InsecureSkipVerify: true,
					})
					if err != nil {
						return fmt.Errorf("failed to get console, %v", err)
					}
				} else {
					server, err = net.Dial("tcp", urlObj.Host)
					if err != nil {
						return fmt.Errorf("failed to get console, %v", err)
					}
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
		} else {
			proxyAddr := "wss://" + turnAddrParts[0] + ":" + sessInfo.AccessPort + "/edgeshell?edgetoken=" + sessInfo.Token
			req.AccessUrl = proxyAddr
			cd.ExecReqSend.Update(ctx, req)
			err = run.proxyRawConn(turnConn)
			if err != nil {
				return err
			}
		}
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

func (s *WebrtcExec) proxyRawConn(turnConn net.Conn) error {
	prd, pwr := io.Pipe()
	go io.Copy(pwr, turnConn)
	err := s.client.Shell(prd, turnConn, turnConn, s.contcmd)
	if err != nil {
		log.DebugLog(log.DebugLevelApi,
			"failed to exec",
			"cmd", s.contcmd, "err", err)
	}
	return err
}
