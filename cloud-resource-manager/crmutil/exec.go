package crmutil

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util/webrtcutil"
	webrtc "github.com/pion/webrtc/v2"
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

func (s *ExecReqHandler) Recv(msg *edgeproto.ExecRequest) {
	// spawn go process so we don't stall notify messages
	go func() {
		err := s.cd.ProcessExecReq(msg)
		if err != nil {
			msg.Err = err.Error()
		}
		s.cd.ExecReqSend.Update(msg)
	}()
}

func (cd *ControllerData) ProcessExecReq(req *edgeproto.ExecRequest) error {
	appInst := edgeproto.AppInst{}
	found := cd.AppInstCache.Get(req.AppInstKey, &appInst)
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
	clusterInst := edgeproto.ClusterInst{}
	found = cd.ClusterInstCache.Get(&appInst.ClusterInstKey, &clusterInst)
	if !found {
		return fmt.Errorf("cluster inst %s not found",
			appInst.ClusterInstKey.GetKeyString())
	}

	run := &WebrtcExec{
		req: req,
	}
	var err error

	run.contcmd, err = cd.platform.GetContainerCommand(&clusterInst, &app, &appInst, req)
	if err != nil {
		return err
	}

	run.client, err = cd.platform.GetPlatformClient(&clusterInst)
	if err != nil {
		return err
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
		log.DebugLog(log.DebugLevelApi,
			"failed to establish peer connection",
			"config", config, "err", err)
		return fmt.Errorf("failed to establish peer connection, %v", err)
	}

	// register handlers
	peerConn.OnDataChannel(run.DataChannel)

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
	return nil
}

type WebrtcExec struct {
	req     *edgeproto.ExecRequest
	client  pc.PlatformClient
	contcmd string
	sin     io.WriteCloser
}

func (s *WebrtcExec) DataChannel(d *webrtc.DataChannel) {
	d.OnOpen(func() {
		wr := webrtcutil.NewDataChanWriter(d)
		prd, pwr := io.Pipe()
		s.sin = pwr

		err := s.client.Shell(prd, wr, wr, "sh", "-c", s.contcmd)
		if err != nil {
			log.DebugLog(log.DebugLevelApi,
				"failed to exec",
				"cmd", s.contcmd, "err", err)
		}
		d.Close()
	})
	d.OnMessage(func(msg webrtc.DataChannelMessage) {
		log.DebugLog(log.DebugLevelApi, "on message",
			"data", msg.Data)
		s.sin.Write(msg.Data)
	})
	d.OnClose(func() {
	})
}
