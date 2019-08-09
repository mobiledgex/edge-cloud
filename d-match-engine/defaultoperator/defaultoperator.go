package defaultoperator

import (
	"fmt"

	simulatedqos "github.com/mobiledgex/edge-cloud-infra/operator-api-gw/defaultoperator/simulated-qos"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	operator "github.com/mobiledgex/edge-cloud/d-match-engine/operator"
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
func (o *OperatorApiGw) Init(operatorName string, servers *operator.OperatorApiGwServers) error {
	log.DebugLog(log.DebugLevelDmereq, "init for default operator", "operatorName", operatorName)
	o.Servers = servers
	return nil
}

func (*OperatorApiGw) VerifyLocation(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply, ckey *dmecommon.CookieKey) (*dme.VerifyLocationReply, error) {
	return nil, fmt.Errorf("Verify Location not supported for this operator")
}

func (*OperatorApiGw) GetQOSPositionKPI(mreq *dme.QosPositionKpiRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	log.DebugLog(log.DebugLevelDmereq, "getting simulated results for operator with no QOS Pos implementation")
	return simulatedqos.GetSimulatedQOSPositionKPI(mreq, getQosSvr)
}
