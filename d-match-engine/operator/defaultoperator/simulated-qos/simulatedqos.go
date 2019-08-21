package simulatedqos

// Provides simulated results for QOS Positioning

import (
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"

	"math/rand"
	"time"
)

// make a random number near the range of the base value (plus or minus 10)
func randomInRange(baseval int) float32 {
	rand.Seed(time.Now().UnixNano())
	min := baseval - 10
	if min < 0 {
		min = 0
	}
	max := baseval + 10
	result := rand.Intn(max-min) + min
	// add some fraction less than 1
	return float32(result) + rand.Float32()
}

// getKpiResults currently just returns some fake results
func getKpiResults(qosres *dme.QosPositionKpiResult) {
	qosres.DluserthroughputMin = randomInRange(1)
	qosres.DluserthroughputMax = randomInRange(100)
	qosres.DluserthroughputAvg = randomInRange(50)
	qosres.UluserthroughputMin = randomInRange(1)
	qosres.UluserthroughputMax = randomInRange(50)
	qosres.UluserthroughputAvg = randomInRange(25)
	qosres.LatencyMin = randomInRange(20)
	qosres.LatencyMax = randomInRange(60)
	qosres.LatencyAvg = randomInRange(40)
}

// GetSimulatedQOSPositionKPI gets fake QOS Position KPIs
func GetSimulatedQOSPositionKPI(mreq *dme.QosPositionRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	log.DebugLog(log.DebugLevelDmereq, "getQosPositionKpi", "request", mreq)

	var mreply dme.QosPositionKpiReply

	mreply.Status = dme.ReplyStatus_RS_SUCCESS

	for _, p := range mreq.Positions {
		pid := p.Positionid
		var qosres dme.QosPositionKpiResult

		qosres.Positionid = pid
		qosres.GpsLocation = p.GpsLocation
		getKpiResults(&qosres)
		log.DebugLog(log.DebugLevelDmereq, "Position", "pid", pid, "qosres", qosres)
		mreply.PositionResults = append(mreply.PositionResults, &qosres)
	}

	getQosSvr.Send(&mreply)
	return nil
}
