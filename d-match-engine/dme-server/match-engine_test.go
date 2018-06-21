package main

import (
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

func TestAddRemove(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelDmeReq)
	setupMatchEngine()
	appInsts := generateAppInsts()

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
	for _, rr := range findCloudletData {
		reply := dme.Match_Engine_Reply{}
		findCloudlet(&rr.req, &reply)
		if !rr.reply.Status {
			assert.False(t, reply.Status)
		} else {
			assert.True(t, reply.Status)
			assert.Equal(t, rr.reply.Uri, reply.Uri)
		}
	}

	// delete all data
	for _, inst := range appInsts {
		removeApp(inst)
	}
	assert.Equal(t, 0, len(tbl.apps))
}

type app struct {
	id        uint64
	name      string
	vers      string
	developer string
}
type cloudlet struct {
	id          uint64
	carrierId   uint64
	carrierName string
	name        string
	uri         string
	location    dme.Loc
}
type findCloudletRR struct {
	req   dme.Match_Engine_Request
	reply dme.Match_Engine_Reply
}

var apps = []app{
	app{
		id:        5000,
		name:      "1000realities",
		vers:      "1.1",
		developer: "1000realities",
	},
	app{
		id:        5005,
		name:      "Pokemon-go",
		vers:      "2.1",
		developer: "Niantic Labs",
	},
	app{
		id:        5006,
		name:      "HarryPotter-go",
		vers:      "1.0",
		developer: "Niantic Labs",
	},
	app{
		id:        5010,
		name:      "Ever",
		vers:      "1.7",
		developer: "Ever.AI",
	},
	app{
		id:        5011,
		name:      "EmptyMatchEngineApp",
		vers:      "1",
		developer: "EmptyMatchEngineApp",
	},
}

var cloudlets = []cloudlet{
	cloudlet{
		id:          111,
		carrierId:   1,
		carrierName: "TDG",
		name:        "Bonn",
		uri:         "10.1.10.1",
		location:    dme.Loc{Lat: 50.7374, Long: 7.0982},
	},
	cloudlet{
		id:          222,
		carrierId:   1,
		carrierName: "TDG",
		name:        "Munich",
		uri:         "11.1.11.1",
		location:    dme.Loc{Lat: 52.7374, Long: 13.4050},
	},
	cloudlet{
		id:          333,
		carrierId:   1,
		carrierName: "TDG",
		name:        "Berlin",
		uri:         "12.1.12.1",
		location:    dme.Loc{Lat: 48.1351, Long: 11.5820},
	},
	cloudlet{
		id:          444,
		carrierId:   3,
		carrierName: "TMUS",
		name:        "San Francisco",
		uri:         "13.1.13.1",
		location:    dme.Loc{Lat: 47.6062, Long: 122.3321},
	},
}

var findCloudletData = []findCloudletRR{
	findCloudletRR{
		req: dme.Match_Engine_Request{
			CarrierName: "TDG",
			GpsLocation: &dme.Loc{Lat: 50.65, Long: 6.341},
			DevName:     "1000realities",
			AppName:     "1000realities",
			AppVers:     "1.1",
		},
		reply: dme.Match_Engine_Reply{
			Uri:              cloudlets[2].uri,
			CloudletLocation: &cloudlets[2].location,
			Status:           true,
		},
	},
	findCloudletRR{
		req: dme.Match_Engine_Request{
			CarrierName: "TDG",
			GpsLocation: &dme.Loc{Lat: 51.65, Long: 9.341},
			DevName:     "1000realities",
			AppName:     "1000realities",
			AppVers:     "1.1",
		},
		reply: dme.Match_Engine_Reply{
			Uri:              cloudlets[1].uri,
			CloudletLocation: &cloudlets[1].location,
			Status:           true,
		},
	},
	findCloudletRR{
		req: dme.Match_Engine_Request{
			CarrierName: "ATT",
			GpsLocation: &dme.Loc{Lat: 52.65, Long: 10.341},
			DevName:     "1000realities",
			AppName:     "1000realities",
			AppVers:     "1.1",
		},
		reply: dme.Match_Engine_Reply{
			Status: false,
		},
	},
}

func makeAppInst(a *app, c *cloudlet) *edgeproto.AppInst {
	inst := edgeproto.AppInst{}
	inst.Key.AppKey.DeveloperKey.Name = a.developer
	inst.Key.AppKey.Name = a.name
	inst.Key.AppKey.Version = a.vers
	inst.Key.CloudletKey.OperatorKey.Name = c.carrierName
	inst.Key.CloudletKey.Name = c.name
	inst.CloudletLoc = c.location
	inst.Uri = c.uri
	return &inst
}

func generateAppInsts() []*edgeproto.AppInst {
	insts := make([]*edgeproto.AppInst, 0)
	for _, c := range cloudlets {
		for _, a := range apps {
			insts = append(insts, makeAppInst(&a, &c))
		}
	}
	return insts
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
