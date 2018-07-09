package main

import (
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/assert"
)

func TestAddRemove(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	setupMatchEngine()
	appInsts := dmetest.GenerateAppInsts()

	tbl := carrierAppTbl

	// add all data, check that number of instances matches
	for _, inst := range appInsts {
		addApp(inst)
	}
	checkAllData(t, appInsts)

	// re-add data, counts should remain unchanged
	for _, inst := range appInsts {
		addApp(inst)
	}
	checkAllData(t, appInsts)

	// delete one data, check new counts
	removeApp(appInsts[0])
	remaining := appInsts[1:]
	checkAllData(t, remaining)

	// test findCloudlet
	for ii, rr := range dmetest.FindCloudletData {
		reply := dme.Match_Engine_Reply{}
		findCloudlet(&rr.Req, &reply)
		assert.Equal(t, rr.Reply.Status, reply.Status, "findCloudletData[%d]", ii)
		if reply.Status == dme.Match_Engine_Reply_FIND_FOUND {
			assert.Equal(t, rr.Reply.Uri, reply.Uri,
				"findCloudletData[%d]", ii)
		}
	}

	// delete all data
	for _, inst := range appInsts {
		removeApp(inst)
	}
	assert.Equal(t, 0, len(tbl.apps))
}

type dummyCarrierApp struct {
	insts map[edgeproto.CloudletKey]struct{}
}

func checkAllData(t *testing.T, appInsts []*edgeproto.AppInst) {
	tbl := carrierAppTbl

	appsCheck := make(map[carrierAppKey]*dummyCarrierApp)
	for _, inst := range appInsts {
		key := carrierAppKey{}
		setCarrierAppKey(inst, &key)
		app, found := appsCheck[key]
		if !found {
			app = &dummyCarrierApp{}
			app.insts = make(map[edgeproto.CloudletKey]struct{})
			appsCheck[key] = app
		}
		app.insts[inst.Key.CloudletKey] = struct{}{}
	}
	assert.Equal(t, len(appsCheck), len(tbl.apps), "Number of carrier apps")
	for k, app := range tbl.apps {
		appChk, found := appsCheck[k]
		assert.True(t, found, "found app %s", k)
		if !found {
			continue
		}
		assert.Equal(t, len(appChk.insts), len(app.insts), "Number of cloudlets")
	}
}
