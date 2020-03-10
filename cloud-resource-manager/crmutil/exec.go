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

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
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
	found := cd.AppInstCache.Get(&req.AppInstKey, &appInst)
	if !found {
		return fmt.Errorf("app inst %s not found",
			req.AppInstKey.GetKeyString())
	}
	app := edgeproto.App{}
	found = cd.AppCache.Get(&req.AppInstKey.AppKey, &app)
	if !found {
		return fmt.Errorf("app %s not found",
			req.AppInstKey.AppKey.GetKeyString())
	}

	consoleUrl := ""
	if req.Console != nil {
		consoleUrl, err = cd.platform.GetConsoleUrl(ctx, &app)
		if err != nil {
			return err
		}
	} else {
		clusterInst := edgeproto.ClusterInst{}
		found = cd.ClusterInstCache.Get(&appInst.Key.ClusterInstKey, &clusterInst)
		if !found {
			return fmt.Errorf("cluster inst %s not found",
				appInst.Key.ClusterInstKey.GetKeyString())
		}

		run.contcmd, err = cd.platform.GetContainerCommand(ctx, &clusterInst, &app, &appInst, req)
		if err != nil {
			return err
		}

		run.client, err = cd.platform.GetPlatformClient(ctx, &clusterInst)
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
			log.DebugLog(log.DebugLevelApi,
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
			log.DebugLog(log.DebugLevelApi,
				"failed to set remote description",
				"offer", offer, "peerConn", peerConn, "err", err)
			return fmt.Errorf("failed to set remote description, %v", err)
		}
		// create answer
		answer, err := peerConn.CreateAnswer(nil)
		if err != nil {
			log.DebugLog(log.DebugLevelApi,
				"failed to create answer",
				"peerConn", peerConn, "err", err)
			return fmt.Errorf("failed to set answer, %v", err)
		}
		// set local description, and starts out UDP listeners
		err = peerConn.SetLocalDescription(answer)
		if err != nil {
			log.DebugLog(log.DebugLevelApi,
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
		log.DebugLog(log.DebugLevelApi, "returning answer")
	} else {
		if req.EdgeTurnAddr == "" {
			return fmt.Errorf("no edgeturn server address specified")
		}

		turnConn, err := tls.Dial("tcp", req.EdgeTurnAddr, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return fmt.Errorf("failed to connect to edgeturn server: %v", err)
		}
		defer turnConn.Close()

		// Fetch session info
		var sessInfo util.SessionInfo
		d := json.NewDecoder(turnConn)
		err = d.Decode(&sessInfo)
		if err != nil {
			return fmt.Errorf("failed to decode session info: %v", err)
		}
		log.DebugLog(log.DebugLevelApi, "recieved session info from edgeturn server", "info", sessInfo)

		if req.Console != nil {
			//TODO: Get proper URL
			consoleUrlParts := strings.Split(consoleUrl, "?token=")
			addrParts := strings.Split(req.EdgeTurnAddr, ":")

			proxyUrl := addrParts[0] + ":" + sessInfo.Port + "/" + consoleUrlParts[len(consoleUrlParts)-1]
			req.Console.Url = proxyUrl

			// Notify controller about the new proxy Console URL
			cd.ExecReqSend.Update(ctx, req)
			run.proxyConsoleConn(consoleUrl, turnConn)
		} else {
			run.proxyRawConn(turnConn)
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

func (s *WebrtcExec) proxyConsoleConn(fromUrl string, toConn io.ReadWriteCloser) {
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

func (s *WebrtcExec) proxyRawConn(toConn io.ReadWriteCloser) {
	go io.Copy(s.sin, toConn)

	prd, pwr := io.Pipe()
	s.sin = pwr

	err := s.client.Shell(prd, toConn, toConn, s.contcmd)
	if err != nil {
		log.DebugLog(log.DebugLevelApi,
			"failed to exec",
			"cmd", s.contcmd, "err", err)
	}
}
