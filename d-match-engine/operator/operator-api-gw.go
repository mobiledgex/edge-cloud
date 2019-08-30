package operator

import (
	"context"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

type OperatorApiGwServers struct {
	VaultAddr string
	QosPosUrl string
	LocVerUrl string
	TokSrvUrl string
}

// OperatorApiGw implements operator specific APIs.
type OperatorApiGw interface {
	// GetOperator Returns the operator name
	GetOperatorName() string
	// Init is called once during startup.
	Init(ctx context.Context, operatorName string, servers *OperatorApiGwServers) error
	// VerifyLocation verifies a client's location against the coordinates provided
	VerifyLocation(ctx context.Context, mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply) error
	// GetLocation gets the client location
	GetLocation(ctx context.Context, mreq *dme.GetLocationRequest, mreply *dme.GetLocationReply) error
	// GetQOSPositionKPI gets QOS KPIs for GPS positions
	GetQOSPositionKPI(req *dme.QosPositionRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error
}
