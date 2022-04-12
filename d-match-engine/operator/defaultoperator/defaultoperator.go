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

package defaultoperator

import (
	"context"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/version"

	operator "github.com/mobiledgex/edge-cloud/d-match-engine/operator"
	simulatedloc "github.com/mobiledgex/edge-cloud/d-match-engine/operator/defaultoperator/simulated-location"
	simulatedqos "github.com/mobiledgex/edge-cloud/d-match-engine/operator/defaultoperator/simulated-qos"

	"github.com/mobiledgex/edge-cloud/log"
)

//OperatorApiGw represents an Operator API Gateway
type OperatorApiGw struct {
	Servers *operator.OperatorApiGwServers
}

func (OperatorApiGw) GetOperatorName() string {
	return "default"
}

// Init is called once during startup.
func (o *OperatorApiGw) Init(operatorName string, servers *operator.OperatorApiGwServers) error {
	log.DebugLog(log.DebugLevelDmereq, "init for default operator", "operatorName", operatorName)
	o.Servers = servers
	return nil
}

func (*OperatorApiGw) VerifyLocation(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply) error {
	return simulatedloc.VerifySimulatedClientLoc(mreq, mreply)
}

func (*OperatorApiGw) GetLocation(mreq *dme.GetLocationRequest, mreply *dme.GetLocationReply) error {
	return simulatedloc.GetSimulatedClientLoc(mreq, mreply)
}

func (*OperatorApiGw) GetQOSPositionKPI(mreq *dme.QosPositionRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	log.DebugLog(log.DebugLevelDmereq, "getting simulated results for operator with no QOS Pos implementation")
	return simulatedqos.GetSimulatedQOSPositionKPI(mreq, getQosSvr)
}

func (*OperatorApiGw) GetVersionProperties() map[string]string {
	return version.BuildProps("DefaultOperator")
}

func (*OperatorApiGw) CreatePrioritySession(ctx context.Context, req *dme.QosPrioritySessionCreateRequest) (*dme.QosPrioritySessionReply, error) {
	log.DebugLog(log.DebugLevelDmereq, "No default implementation for CreatePrioritySession()")
	return nil, nil
}

func (*OperatorApiGw) DeletePrioritySession(ctx context.Context, req *dme.QosPrioritySessionDeleteRequest) (*dme.QosPrioritySessionDeleteReply, error) {
	log.DebugLog(log.DebugLevelDmereq, "No default implementation for DeletePrioritySession()")
	return nil, nil
}
