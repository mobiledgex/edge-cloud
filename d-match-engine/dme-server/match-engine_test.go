package main

import (
	"testing"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestAddRemove(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	setupMatchEngine()
	appInsts := dmetest.GenerateAppInsts()

	tbl := dmeAppTbl

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
	serv := server{}

	// test findCloudlet
	for ii, rr := range dmetest.FindCloudletData {
		ctx := dmecommon.PeerContext(context.Background(), "127.0.0.1", 123)

		regReply, err := serv.RegisterClient(ctx, &rr.Reg)
		assert.Nil(t, err, "register client")

		// Since we're directly calling functions, we end up
		// bypassing the interceptor which sets up the cookie key.
		// So set it on the context manually.
		ckey, err := dmecommon.VerifyCookie(regReply.SessionCookie)
		assert.Nil(t, err, "verify cookie")
		ctx = dmecommon.NewCookieContext(ctx, ckey)

		reply, err := serv.FindCloudlet(ctx, &rr.Req)
		assert.Nil(t, err, "find cloudlet")
		assert.Equal(t, rr.Reply.Status, reply.Status, "findCloudletData[%d]", ii)
		if reply.Status == dme.FindCloudletReply_FIND_FOUND {
			assert.Equal(t, rr.Reply.FQDN, reply.FQDN,
				"findCloudletData[%d]", ii)
		}
	}

	// delete all data
	for _, inst := range appInsts {
		removeApp(inst)
	}
	assert.Equal(t, 0, len(tbl.apps))
}

type dummyDmeApp struct {
	insts map[edgeproto.CloudletKey]struct{}
}

func checkAllData(t *testing.T, appInsts []*edgeproto.AppInst) {
	tbl := dmeAppTbl

	appsCheck := make(map[edgeproto.AppKey]*dummyDmeApp)
	for _, inst := range appInsts {
		appkey := inst.Key.AppKey
		app, found := appsCheck[appkey]
		if !found {
			app = &dummyDmeApp{}
			app.insts = make(map[edgeproto.CloudletKey]struct{})
			appsCheck[appkey] = app
		}
		app.insts[inst.Key.CloudletKey] = struct{}{}
	}
	assert.Equal(t, len(appsCheck), len(tbl.apps), "Number of carrier apps")
	for k, app := range tbl.apps {
		for _, c := range app.carriers {
			appChk, found := appsCheck[k]
			assert.True(t, found, "found app %s", k)
			if !found {
				continue
			}
			assert.Equal(t, len(appChk.insts), len(c.insts), "Number of cloudlets")
		}
	}
}
