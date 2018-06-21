package main

import (
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

func TestVerifyLoc(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelDmeReq)
	setupMatchEngine()
	appInsts := generateAppInsts()

	// add all data
	for _, inst := range appInsts {
		addApp(inst)
	}
	// test verify location
	for _, rr := range verifyLocData {
		reply := dme.Match_Engine_Loc_Verify{}
		VerifyClientLoc(&rr.req, &reply)
		assert.Equal(t, rr.reply.GpsLocationStatus, reply.GpsLocationStatus)
	}
}

type verifyLocRR struct {
	req   dme.Match_Engine_Request
	reply dme.Match_Engine_Loc_Verify
}

var verifyLocData = []verifyLocRR{
	verifyLocRR{
		req: dme.Match_Engine_Request{
			CarrierName: "TDG",
			GpsLocation: &dme.Loc{Lat: 50.73, Long: 7.1},
			DevName:     "1000realities",
			AppName:     "1000realities",
			AppVers:     "1.1",
		},
		reply: dme.Match_Engine_Loc_Verify{
			GpsLocationStatus: 1,
		},
	},
	verifyLocRR{
		req: dme.Match_Engine_Request{
			CarrierName: "TDG",
			GpsLocation: &dme.Loc{Lat: 52.65, Long: 12.341},
			DevName:     "1000realities",
			AppName:     "1000realities",
			AppVers:     "1.1",
		},
		reply: dme.Match_Engine_Loc_Verify{
			GpsLocationStatus: 3,
		},
	},
	verifyLocRR{
		req: dme.Match_Engine_Request{
			CarrierName: "ATT",
			GpsLocation: &dme.Loc{Lat: 52.65, Long: 10.341},
			DevName:     "1000realities",
			AppName:     "1000realities",
			AppVers:     "1.1",
		},
		reply: dme.Match_Engine_Loc_Verify{
			GpsLocationStatus: 0,
		},
	},
}
