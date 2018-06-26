package main

import (
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/assert"
)

func TestVerifyLoc(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	setupMatchEngine()
	appInsts := dmetest.GenerateAppInsts()

	// add all data
	for _, inst := range appInsts {
		addApp(inst)
	}
	// test verify location
	for ii, rr := range dmetest.VerifyLocData {
		reply := dme.Match_Engine_Loc_Verify{}
		VerifyClientLoc(&rr.Req, &reply)
		assert.Equal(t, rr.Reply.GpsLocationStatus, reply.GpsLocationStatus,
			"VerifyLocData[%d]", ii)
	}
}
