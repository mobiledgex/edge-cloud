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

package operator

import (
	"context"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

type OperatorApiGwServers struct {
	VaultAddr  string
	QosPosUrl  string
	LocVerUrl  string
	TokSrvUrl  string
	QosSesAddr string
}

type QosSessionEndpoints struct {
}

// OperatorApiGw implements operator specific APIs.
type OperatorApiGw interface {
	// GetOperator Returns the operator name
	GetOperatorName() string
	// GetVersionProperties returns properties related to the ApiGw version
	GetVersionProperties() map[string]string
	// Init is called once during startup.
	Init(operatorName string, servers *OperatorApiGwServers) error
	// VerifyLocation verifies a client's location against the coordinates provided
	VerifyLocation(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply) error
	// GetLocation gets the client location
	GetLocation(mreq *dme.GetLocationRequest, mreply *dme.GetLocationReply) error
	// GetQOSPositionKPI gets QOS KPIs for GPS positions
	GetQOSPositionKPI(req *dme.QosPositionRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error
	// CreatePrioritySession requests either stable latency or throughput for a client session
	CreatePrioritySession(ctx context.Context, req *dme.QosPrioritySessionCreateRequest) (*dme.QosPrioritySessionReply, error)
	// DeletePrioritySession removes a previously created priority session
	DeletePrioritySession(ctx context.Context, req *dme.QosPrioritySessionDeleteRequest) (*dme.QosPrioritySessionDeleteReply, error)
}
