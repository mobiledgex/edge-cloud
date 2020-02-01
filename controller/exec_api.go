package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
)

type ExecApi struct {
	requests map[string]*ExecReq
	mux      sync.Mutex
}

type ExecReq struct {
	req  edgeproto.ExecRequest
	done chan bool
}

const (
	// For K8s/Docker based Apps
	ShortTimeout = edgeproto.Duration(6 * time.Second)
	// For VM based Apps
	LongTimeout = edgeproto.Duration(20 * time.Second)
)

var execApi = ExecApi{}
var execRequestSendMany *notify.ExecRequestSendMany

func InitExecApi() {
	execApi.requests = make(map[string]*ExecReq)
	execRequestSendMany = notify.NewExecRequestSendMany()
}

func (s *ExecApi) getApp(req *edgeproto.ExecRequest, app *edgeproto.App) error {
	if req.AppInstKey.ClusterInstKey.Developer == "" {
		req.AppInstKey.ClusterInstKey.Developer = req.AppInstKey.AppKey.DeveloperKey.Name
	}
	if !appInstApi.HasKey(&req.AppInstKey) {
		return req.AppInstKey.NotFoundError()
	}
	if !appApi.Get(&req.AppInstKey.AppKey, app) {
		return req.AppInstKey.AppKey.NotFoundError()
	}
	return nil
}

func (s *ExecApi) ViewLogs(ctx context.Context, req *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	if req.Log == nil {
		// defaults
		req.Log = &edgeproto.ViewLog{}
	}
	app := edgeproto.App{}
	if err := s.getApp(req, &app); err != nil {
		return nil, err
	}
	req.Timeout = ShortTimeout
	// Be very careful about validating string input. These arguments
	// will be passed to the command line in the VM, which user should
	// not have access to.
	if app.Deployment == cloudcommon.AppDeploymentTypeDocker {
		if req.ContainerId != "" && !util.ValidDockerName(req.ContainerId) {
			return nil, fmt.Errorf("Invalid docker container name")
		}
	} else if app.Deployment == cloudcommon.AppDeploymentTypeKubernetes {
		if req.ContainerId != "" && !util.ValidKubernetesName(req.ContainerId) {
			return nil, fmt.Errorf("Invalid kubernetes container name")
		}
	} else {
		return nil, fmt.Errorf("Unsupported deployment type")
	}
	if req.Log.Since != "" {
		_, err := time.ParseDuration(req.Log.Since)
		if err != nil {
			_, err = time.Parse(time.RFC3339, req.Log.Since)
		}
		if err != nil {
			return nil, fmt.Errorf("Unable to parse Since field as duration or RFC3339 formatted time")
		}
	}
	return s.doExchange(ctx, req)
}

func (s *ExecApi) RunCommand(ctx context.Context, req *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	cmd := req.Cmd
	if cmd == nil {
		return nil, fmt.Errorf("No run command specified")
	}
	app := edgeproto.App{}
	if err := s.getApp(req, &app); err != nil {
		return nil, err
	}
	req.Timeout = ShortTimeout
	if app.Deployment == cloudcommon.AppDeploymentTypeVM {
		req.Timeout = LongTimeout
		if cmd.Command != "" || req.ContainerId != "" {
			return nil, fmt.Errorf("invalid argument, command/containerid is not supported for VM based Apps")
		}
		cmd.Console = true
	} else {
		if cmd.Command == "" {
			return nil, fmt.Errorf("command argument is mandatory for %s based Apps", app.Deployment)
		}
	}
	return s.doExchange(ctx, req)
}

func (s *ExecApi) doExchange(ctx context.Context, req *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	// Forward the offer.
	// Currently we don't know which controller has the CRM connected
	// (or if it's even present), so just broadcast to all.
	err := controllerApi.RunJobs(func(arg interface{}, addr string) error {
		var err error
		var reply *edgeproto.ExecRequest
		if addr == *externalApiAddr {
			// local node
			reply, err = s.SendLocalRequest(ctx, req)
		} else {
			// connect to remote node
			conn, err := ControllerConnect(addr)
			if err != nil {
				return err
			}
			defer conn.Close()

			cmd := edgeproto.NewExecApiClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), req.Timeout.TimeDuration()+2)
			defer cancel()
			reply, err = cmd.SendLocalRequest(ctx, req)
		}
		if err == nil && reply != nil {
			req.Answer = reply.Answer
			req.Err = reply.Err
			if req.Cmd != nil && reply.Cmd != nil {
				req.Cmd.ConsoleUrl = reply.Cmd.ConsoleUrl
			}
		}
		return nil
	}, nil)
	if err != nil {
		return nil, err
	}
	if req.Err != "" {
		return nil, fmt.Errorf("%s", req.Err)
	}
	if req.Answer == "" {
		return nil, fmt.Errorf("no one answered the offer")
	}
	log.DebugLog(log.DebugLevelApi, "ExecRequest answered", "req", req)
	return req, nil
}

// sendLocalRequest sends the request over the notify framework
// to all CRM clients. We then wait for a CRM client with the
// AppInst to reply with an answer to the offer.
func (s *ExecApi) SendLocalRequest(ctx context.Context, req *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	s.mux.Lock()
	sr := &ExecReq{
		req:  *req,
		done: make(chan bool, 1),
	}
	s.requests[req.Offer] = sr
	s.mux.Unlock()

	// SendMany filters based on CloudletKey, so will not send if
	// the attached CRM(s) aren't for that CloudletKey. This will
	// return the number of messages sent. We don't need to refcount
	// multiple replies because there should only be one CRM in the
	// system that matches the CloudletKey, so there should only ever
	// be one reply.
	var err error
	count := execRequestSendMany.Update(ctx, req)
	if count == 0 {
		err = fmt.Errorf("no matching CRMs")
	} else {
		// wait for reply or timeout
		select {
		case <-sr.done:
		case <-time.After(req.Timeout.TimeDuration()):
			err = fmt.Errorf("ExecRequest timed out")
		}
	}
	s.mux.Lock()
	delete(s.requests, req.Offer)
	s.mux.Unlock()

	if err != nil {
		return nil, err
	}
	// if we got an answer, it will be in the request struct
	return &sr.req, nil
}

// Receive message from notify framework
func (s *ExecApi) Recv(ctx context.Context, msg *edgeproto.ExecRequest) {
	s.mux.Lock()
	defer s.mux.Unlock()
	sr, ok := s.requests[msg.Offer]
	if !ok {
		log.DebugLog(log.DebugLevelApi, "unregistered ExecRequest recv", "msg", msg)
		return
	}
	sr.req.Answer = msg.Answer
	sr.req.Err = msg.Err
	if sr.req.Cmd != nil && msg.Cmd != nil {
		sr.req.Cmd.ConsoleUrl = msg.Cmd.ConsoleUrl
	}
	sr.done <- true
}
