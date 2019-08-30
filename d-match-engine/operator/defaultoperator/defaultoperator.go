package defaultoperator

import (
	"context"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"

	operator "github.com/mobiledgex/edge-cloud/d-match-engine/operator"
	simulatedloc "github.com/mobiledgex/edge-cloud/d-match-engine/operator/defaultoperator/simulated-location"
	simulatedqos "github.com/mobiledgex/edge-cloud/d-match-engine/operator/defaultoperator/simulated-qos"

	"github.com/mobiledgex/edge-cloud/log"
)

//OperatorApiGw respresent an Operator API Gateway
type OperatorApiGw struct {
	Servers *operator.OperatorApiGwServers
}

func (OperatorApiGw) GetOperatorName() string {
	return "default"
}

// Init is called once during startup.
func (o *OperatorApiGw) Init(ctx context.Context, operatorName string, servers *operator.OperatorApiGwServers) error {
	log.SpanLog(ctx, log.DebugLevelDmereq, "init for default operator", "operatorName", operatorName)
	o.Servers = servers
	return nil
}

func (*OperatorApiGw) VerifyLocation(ctx context.Context, mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply) error {
	return simulatedloc.VerifySimulatedClientLoc(mreq, mreply)
}

func (*OperatorApiGw) GetLocation(ctx context.Context, mreq *dme.GetLocationRequest, mreply *dme.GetLocationReply) error {
	return simulatedloc.GetSimulatedClientLoc(mreq, mreply)
}

func (*OperatorApiGw) GetQOSPositionKPI(mreq *dme.QosPositionRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	log.DebugLog(log.DebugLevelDmereq, "getting simulated results for operator with no QOS Pos implementation")
	return simulatedqos.GetSimulatedQOSPositionKPI(mreq, getQosSvr)
}
