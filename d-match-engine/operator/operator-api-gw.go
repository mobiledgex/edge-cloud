package operator

import (
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

// OperatorApiGw implements operator specific APIs.
type OperatorApiGw interface {
	// GetOperator Returns the operator name
	GetOperatorName() string
	// Init is called once during startup.
	Init(operatorName, vaultAddr, qosPosUrl, locVerUrl, tokSrvUrl string) error
	// VerifyLocation verifies a clien's location against the coordinates provided
	VerifyLocation(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply, ckey *dmecommon.CookieKey) (*dme.VerifyLocationReply, error)
	// GetQOSPositionKPI gets QOS KPIs for GPS positions
	GetQOSPositionKPI(req *dme.QosPositionKpiRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error
}
