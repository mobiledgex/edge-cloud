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
	ssh "github.com/mobiledgex/golang-ssh"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/xtaci/smux"
)

const ExecRequestIgnoredPlatformInactive = "ExecRequestIgnoredPlatformInactive"

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
			if err.Error() == ExecRequestIgnoredPlatformInactive {
				// send nothing in response as the controller only looks for one response
				return
			}
			msg.Err = err.Error()
		}
		s.cd.ExecReqSend.Update(ctx, msg)
	}()
}

type EnvoyContainerVersion struct {
	ContainerName string
	EnvoyVersion  string
	Error         string
}

type RootLBEnvoyVersion struct {
	NodeType        string
	NodeName        string
	EnvoyContainers []EnvoyContainerVersion
}

func (cd *ControllerData) GetClusterEnvoyVersion(ctx context.Context, req *edgeproto.DebugRequest) string {
	clusterInsts := []edgeproto.ClusterInst{}
	cd.ClusterInstCache.Mux.Lock()
	for _, v := range cd.ClusterInstCache.Objs {
		clusterInsts = append(clusterInsts, *v.Obj)
	}
	cd.ClusterInstCache.Mux.Unlock()
	nodes, err := cd.platform.ListCloudletMgmtNodes(ctx, clusterInsts, nil)
	if err != nil {
		return fmt.Sprintf("unable to get list of cluster nodes, %v", err)
	}
	if len(nodes) == 0 {
		return fmt.Sprintf("no nodes found")
	}
	nodeVersions := []RootLBEnvoyVersion{}
	for _, node := range nodes {
		if !strings.Contains(node.Type, "rootlb") {
			continue
		}
		client, err := cd.platform.GetNodePlatformClient(ctx, &node)
		if err != nil {
			return fmt.Sprintf("failed to get ssh client for node %s, %v", node.Name, err)
		}
		out, err := client.Output(`docker ps --format "{{.Names}}" --filter name="^envoy"`)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "failed to find envoy containers on rootlb", "rootlb", node, "err", err, "out", out)
			return fmt.Sprintf("failed to find envoy containers on rootlb %s, %v", node.Name, err)
		}
		nodeVersion := RootLBEnvoyVersion{
			NodeType: node.Type,
			NodeName: node.Name,
		}
		for _, name := range strings.Split(out, "\n") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			envoyContainerVers := EnvoyContainerVersion{
				ContainerName: name,
			}
			out, err := client.Output(fmt.Sprintf("docker exec %s envoy --version", name))
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "failed to find envoy container version on rootlb", "rootlb", node, "container", name, "err", err, "out", out)
				envoyContainerVers.Error = err.Error()
				nodeVersion.EnvoyContainers = append(nodeVersion.EnvoyContainers, envoyContainerVers)
				continue
			}
			version := strings.TrimSpace(out)
			envoyContainerVers.EnvoyVersion = version
			nodeVersion.EnvoyContainers = append(nodeVersion.EnvoyContainers, envoyContainerVers)
		}
		nodeVersions = append(nodeVersions, nodeVersion)
	}
	out, err := json.Marshal(nodeVersions)
	if err != nil {
		return fmt.Sprintf("Failed to marshal node versions: %s, %v", string(out), err)
	}
	return string(out)
}

func (cd *ControllerData) ProcessExecReq(ctx context.Context, req *edgeproto.ExecRequest) (reterr error) {
	var err error
	log.SpanLog(ctx, log.DebugLevelApi, "ProcessExecReq", "req", req, "PlatformInstanceActive", cd.highAvailabilityManager.PlatformInstanceActive)

	if !cd.highAvailabilityManager.PlatformInstanceActive {
		return fmt.Errorf(ExecRequestIgnoredPlatformInactive)
	}
	run := &RunExec{
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
		req.Console.Url, err = cd.platform.GetConsoleUrl(ctx, &app, &appInst)
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
		clusterInsts := []edgeproto.ClusterInst{}
		cd.ClusterInstCache.Mux.Lock()
		for _, v := range cd.ClusterInstCache.Objs {
			clusterInsts = append(clusterInsts, *v.Obj)
		}
		cd.ClusterInstCache.Mux.Unlock()
		vmAppInsts := []edgeproto.AppInst{}
		cd.AppInstCache.Mux.Lock()
		for _, v := range cd.AppInstCache.Objs {
			appObj := edgeproto.App{}
			found := cd.AppCache.Get(&v.Obj.Key.AppKey, &appObj)
			if found && appObj.Deployment == cloudcommon.DeploymentTypeVM {
				vmAppInsts = append(vmAppInsts, *v.Obj)
			}
		}
		cd.AppInstCache.Mux.Unlock()
		nodes, err := cd.platform.ListCloudletMgmtNodes(ctx, clusterInsts, vmAppInsts)
		if err != nil {
			return fmt.Errorf("unable to get list of cloudlet mgmt nodes, %v", err)
		}
		if len(nodes) == 0 {
			return fmt.Errorf("no nodes found")
		}
		accessNode := req.Cmd.CloudletMgmtNode
		matchedNodes := []edgeproto.CloudletMgmtNode{}
		for _, node := range nodes {
			// filter by specified node/type.
			// blank means match any.
			if accessNode.Type != "" && accessNode.Type != node.Type {
				continue
			}
			if accessNode.Name != "" && accessNode.Name != node.Name {
				continue
			}
			matchedNodes = append(matchedNodes, node)
		}
		if len(matchedNodes) == 0 {
			return fmt.Errorf("unable to find specified cloudlet mgmt node, list of valid nodes: %v", nodes)
		} else if len(matchedNodes) > 1 {
			return fmt.Errorf("too many nodes matched, please specify type and name from: %v", matchedNodes)
		}
		accessNode = &matchedNodes[0]

		run.contcmd = "bash"
		if req.Cmd.Command != "" {
			run.contcmd = req.Cmd.Command
		}
		run.client, err = cd.platform.GetNodePlatformClient(ctx, accessNode)
		if err != nil {
			return err
		}
		execReqType = cloudcommon.ExecReqShell
	} else {
		execReqType = cloudcommon.ExecReqShell
		clusterInst := edgeproto.ClusterInst{}
		found := cd.ClusterInstCache.Get(appInst.ClusterInstKey(), &clusterInst)
		if !found {
			return fmt.Errorf("cluster inst %s not found",
				appInst.ClusterInstKey().GetKeyString())
		}

		run.contcmd, err = cd.platform.GetContainerCommand(ctx, &clusterInst, &app, &appInst, req)
		if err != nil {
			return err
		}

		clientType := cloudcommon.GetAppClientType(&app)
		run.client, err = cd.platform.GetClusterPlatformClient(ctx, &clusterInst, clientType)
		if err != nil {
			return err
		}
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

	replySent := false
	// if ExecRequest reply is already sent, we can't send any error back to the
	// client via the ExecRequest. Instead we'll need to write it to the
	// turn connection.
	defer func() {
		if reterr != nil && replySent {
			turnConn.Write([]byte(reterr.Error()))
		}
	}()
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
		replySent = true
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
		replySent = true
		err = run.proxyRawConn(turnConn)
		if err != nil {
			return err
		}
	}

	return nil
}

type RunExec struct {
	req     *edgeproto.ExecRequest
	client  ssh.Client
	contcmd string
	sin     io.WriteCloser
}

func (s *RunExec) proxyRawConn(turnConn net.Conn) error {
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
