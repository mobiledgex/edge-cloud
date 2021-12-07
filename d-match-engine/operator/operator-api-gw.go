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
	CreatePrioritySession(ctx context.Context, priorityType string, ueAddr string, asAddr string, asPort string, protocol string, qos string, duration int64) (string, error)
	// DeletePrioritySession removes a previously created priority session
	DeletePrioritySession(ctx context.Context, priorityType string, sessionId string) error
	// LookupQosParm looks up the QOS API parameter values for each QosSessionProfile enum value.
	LookupQosParm(qos string) string
}
