// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"github.com/segmentio/ksuid"
)

type ExecApi struct {
	all      *AllApis
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
	LongTimeout = edgeproto.Duration(60 * time.Second)
)

var execRequestSendMany *notify.ExecRequestSendMany

func NewExecApi(all *AllApis) *ExecApi {
	execApi := ExecApi{}
	execApi.all = all
	execApi.requests = make(map[string]*ExecReq)
	execRequestSendMany = notify.NewExecRequestSendMany()
	return &execApi
}

func (s *ExecApi) getApp(ctx context.Context, req *edgeproto.ExecRequest, app *edgeproto.App) error {
	SetAppInstKeyDefaults(ctx, &req.AppInstKey, s.all)
	if !s.all.appApi.Get(&req.AppInstKey.AppKey, app) {
		return req.AppInstKey.AppKey.NotFoundError()
	}
	if !s.all.appInstApi.HasKey(&req.AppInstKey) {
		return req.AppInstKey.NotFoundError()
	}
	return nil
}

func ValidateContainerName(deployment, name string) error {
	if deployment == cloudcommon.DeploymentTypeDocker {
		if name != "" && !util.ValidDockerName(name) {
			return fmt.Errorf("Invalid docker container name")
		}
	} else if deployment == cloudcommon.DeploymentTypeKubernetes {
		if name != "" {
			err := util.ValidK8SContainerName(name)
			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("Command not supported for %s deployments", deployment)
	}
	return nil
}

func (s *ExecApi) ShowLogs(ctx context.Context, req *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	if req.Log == nil {
		// defaults
		req.Log = &edgeproto.ShowLog{}
	}
	app := edgeproto.App{}
	if err := s.getApp(ctx, req, &app); err != nil {
		return nil, err
	}
	req.Timeout = ShortTimeout
	// Be very careful about validating string input. These arguments
	// will be passed to the command line in the VM, which user should
	// not have access to.
	if err := ValidateContainerName(app.Deployment, req.ContainerId); err != nil {
		return nil, err
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
	if err := s.getApp(ctx, req, &app); err != nil {
		return nil, err
	}
	if app.Deployment == cloudcommon.DeploymentTypeVM {
		return nil, fmt.Errorf("RunCommand not available for VM deployments, use RunConsole instead")
	}
	// Be very careful about validating string input. These arguments
	// will be passed to the command line in the VM, which user should
	// not have access to.
	if err := ValidateContainerName(app.Deployment, req.ContainerId); err != nil {
		return nil, err
	}
	req.Timeout = ShortTimeout
	if cmd.Command == "" {
		return nil, fmt.Errorf("command argument required")
	}
	return s.doExchange(ctx, req)
}

func (s *ExecApi) AccessCloudlet(ctx context.Context, req *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	cmd := req.Cmd
	if cmd == nil {
		return nil, fmt.Errorf("No run command specified")
	}
	if cmd.CloudletMgmtNode == nil {
		return nil, fmt.Errorf("No cloudlet mgmt node specified")
	}
	req.Timeout = ShortTimeout
	return s.doExchange(ctx, req)
}

func (s *ExecApi) RunConsole(ctx context.Context, req *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	req.Cmd = nil
	req.Log = nil
	req.Console = &edgeproto.RunVMConsole{}

	app := edgeproto.App{}
	if err := s.getApp(ctx, req, &app); err != nil {
		return nil, err
	}
	req.Timeout = LongTimeout
	if app.Deployment != cloudcommon.DeploymentTypeVM {
		return nil, fmt.Errorf("RunConsole only available for VM deployments, use RunCommand instead")
	}
	return s.doExchange(ctx, req)
}

func (s *ExecApi) doExchange(ctx context.Context, req *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	// Make sure EdgeTurn Server Address is present
	if *edgeTurnAddr == "" {
		return nil, fmt.Errorf("EdgeTurn server address is required to run commands")
	}
	req.EdgeTurnAddr = *edgeTurnAddr
	reqId := ksuid.New()
	req.Offer = reqId.String()
	// Increase timeout, as for EdgeTurn based implemention,
	// CRM connects to EdgeTurn server. Hence it can take some
	// time to reply back
	req.Timeout = LongTimeout
	// Forward the offer.
	// Currently we don't know which controller has the CRM connected
	// (or if it's even present), so just broadcast to all.
	err := s.all.controllerApi.RunJobs(func(arg interface{}, addr string) error {
		var err error
		var reply *edgeproto.ExecRequest
		if addr == *externalApiAddr {
			// local node
			reply, err = s.SendLocalRequest(ctx, req)
		} else {
			// connect to remote node
			conn, err := ControllerConnect(ctx, addr)
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
			if req.Console != nil && reply.Console != nil {
				req.Console.Url = reply.Console.Url
			}
			req.AccessUrl = reply.AccessUrl
		}
		return nil
	}, nil)
	if err != nil {
		return nil, err
	}
	if req.Err != "" {
		return nil, fmt.Errorf("%s", req.Err)
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
func (s *ExecApi) RecvExecRequest(ctx context.Context, msg *edgeproto.ExecRequest) {
	s.mux.Lock()
	defer s.mux.Unlock()
	sr, ok := s.requests[msg.Offer]
	if !ok {
		log.DebugLog(log.DebugLevelApi, "unregistered ExecRequest recv", "msg", msg, "requests", s.requests)
		return
	}
	sr.req.Answer = msg.Answer
	sr.req.Err = msg.Err
	if sr.req.Console != nil && msg.Console != nil {
		sr.req.Console.Url = msg.Console.Url
	}
	sr.req.AccessUrl = msg.AccessUrl
	sr.done <- true
}
