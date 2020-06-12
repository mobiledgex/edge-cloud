package main

import (
	"testing"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestAddRemove(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	span := log.SpanFromContext(ctx)

	dmecommon.SetupMatchEngine()
	InitAppInstClients()
	setupJwks()
	apps := dmetest.GenerateApps()
	appInsts := dmetest.GenerateAppInsts()
	cloudlets := dmetest.GenerateClouldlets()

	tbl := dmecommon.DmeAppTbl

	// Add cloudlets first as we check the state via cloudlets
	for _, cloudlet := range cloudlets {
		dmecommon.SetInstStateForCloudlet(ctx, cloudlet)
	}

	// add all data, check that number of instances matches
	for _, inst := range apps {
		dmecommon.AddApp(ctx, inst)
	}
	for _, inst := range appInsts {
		dmecommon.AddAppInst(ctx, inst)
	}
	checkAllData(t, appInsts)

	// re-add data, counts should remain unchanged
	for _, inst := range appInsts {
		dmecommon.AddAppInst(ctx, inst)
	}
	checkAllData(t, appInsts)

	// delete one data, check new counts
	dmecommon.RemoveAppInst(ctx, appInsts[0])
	remaining := appInsts[1:]
	checkAllData(t, remaining)
	serv := server{}

	// test findCloudlet
	for ii, rr := range dmetest.FindCloudletData {
		ctx := dmecommon.PeerContext(context.Background(), "127.0.0.1", 123, span)

		regReply, err := serv.RegisterClient(ctx, &rr.Reg)
		assert.Nil(t, err, "register client")

		// Since we're directly calling functions, we end up
		// bypassing the interceptor which sets up the cookie key.
		// So set it on the context manually.
		ckey, err := dmecommon.VerifyCookie(ctx, regReply.SessionCookie)
		assert.Nil(t, err, "verify cookie")
		// verify that UUID in the response is a new value if it was empty in the request
		if rr.Reg.UniqueId == "" {
			assert.NotEqual(t, regReply.UniqueId, "")
			assert.NotEqual(t, regReply.UniqueIdType, "")
			// should match what's in the cookie
			assert.Equal(t, regReply.UniqueId, ckey.UniqueId)
			assert.Equal(t, regReply.UniqueIdType, ckey.UniqueIdType)
		} else {
			// If it was not empty cookie should have the uuid from the register
			assert.Equal(t, rr.Reg.UniqueId, ckey.UniqueId)
			assert.Equal(t, rr.Reg.UniqueIdType, ckey.UniqueIdType)
		}
		ctx = dmecommon.NewCookieContext(ctx, ckey)
		// Make sure we get the statsKey value filled in
		call := ApiStatCall{}
		ctx = context.WithValue(ctx, dmecommon.StatKeyContextKey, &call.key)

		reply, err := serv.FindCloudlet(ctx, &rr.Req)
		assert.Nil(t, err, "find cloudlet")
		assert.Equal(t, rr.Reply.Status, reply.Status, "findCloudletData[%d]", ii)
		if reply.Status == dme.FindCloudletReply_FIND_FOUND {
			assert.Equal(t, rr.Reply.Fqdn, reply.Fqdn,
				"findCloudletData[%d]", ii)
			// Check the filled in cloudlet details
			assert.Equal(t, rr.ReplyCarrier,
				call.key.CloudletFound.Organization, "findCloudletData[%d]", ii)
			assert.Equal(t, rr.ReplyCloudlet,
				call.key.CloudletFound.Name, "findCloudletData[%d]", ii)
		}
	}
	// Check Platform Devices register UUID
	reg := dmetest.DeviceData[0]
	// Both or none should be set
	reg.UniqueId = "123"
	reg.UniqueIdType = ""
	ctx = dmecommon.PeerContext(context.Background(), "127.0.0.1", 123, span)
	regReply, err := serv.RegisterClient(ctx, &reg)
	assert.NotNil(t, err, "register client")
	assert.Contains(t, err.Error(), "Both, or none of UniqueId and UniqueIdType should be set")
	reg.UniqueId = ""
	reg.UniqueIdType = "typeOnly"
	regReply, err = serv.RegisterClient(ctx, &reg)
	assert.NotNil(t, err, "register client")
	assert.Contains(t, err.Error(), "Both, or none of UniqueId and UniqueIdType should be set")
	// Reset UUID to empty strings
	reg.UniqueId = ""
	reg.UniqueIdType = ""
	regReply, err = serv.RegisterClient(ctx, &reg)
	assert.Nil(t, err, "register client")
	ckey, err := dmecommon.VerifyCookie(ctx, regReply.SessionCookie)
	assert.Nil(t, err, "verify cookie")
	// verify that UUID type is the platform one
	assert.Equal(t, reg.OrgName+":"+reg.AppName, regReply.UniqueIdType)
	// should match what's in the cookie
	assert.Equal(t, regReply.UniqueId, ckey.UniqueId)
	assert.Equal(t, regReply.UniqueIdType, ckey.UniqueIdType)

	// disable one cloudlet and check the newly found cloudlet
	cloudletInfo := cloudlets[2]
	cloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_UNKNOWN
	dmecommon.SetInstStateForCloudlet(ctx, cloudletInfo)
	ctx = dmecommon.PeerContext(context.Background(), "127.0.0.1", 123, span)

	regReply, err = serv.RegisterClient(ctx, &dmetest.DisabledCloudletRR.Reg)
	assert.Nil(t, err, "register client")
	ckey, err = dmecommon.VerifyCookie(ctx, regReply.SessionCookie)
	assert.Nil(t, err, "verify cookie")
	ctx = dmecommon.NewCookieContext(ctx, ckey)

	reply, err := serv.FindCloudlet(ctx, &dmetest.DisabledCloudletRR.Req)
	assert.Nil(t, err, "find cloudlet")
	assert.Equal(t, dmetest.DisabledCloudletRR.Reply.Status, reply.Status)
	assert.Equal(t, dmetest.DisabledCloudletRR.Reply.Fqdn, reply.Fqdn)
	// re-enable and check that the results is now what original findCloudlet[3] is
	cloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_READY
	dmecommon.SetInstStateForCloudlet(ctx, cloudletInfo)
	reply, err = serv.FindCloudlet(ctx, &dmetest.DisabledCloudletRR.Req)
	assert.Nil(t, err, "find cloudlet")
	assert.Equal(t, dmetest.FindCloudletData[3].Reply.Status, reply.Status)
	assert.Equal(t, dmetest.FindCloudletData[3].Reply.Fqdn, reply.Fqdn)

	// Change the health check status of the appInst and get check the results
	appInst := dmetest.MakeAppInst(&dmetest.Apps[0], &dmetest.Cloudlets[2])
	appInst.HealthCheck = edgeproto.HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE
	dmecommon.AddAppInst(ctx, appInst)
	reply, err = serv.FindCloudlet(ctx, &dmetest.DisabledCloudletRR.Req)
	assert.Nil(t, err, "find cloudlet")
	assert.Equal(t, dmetest.DisabledCloudletRR.Reply.Status, reply.Status)
	assert.Equal(t, dmetest.DisabledCloudletRR.Reply.Fqdn, reply.Fqdn)
	// reset and check the one that we get is returned
	appInst.HealthCheck = edgeproto.HealthCheck_HEALTH_CHECK_OK
	dmecommon.AddAppInst(ctx, appInst)
	reply, err = serv.FindCloudlet(ctx, &dmetest.DisabledCloudletRR.Req)
	assert.Nil(t, err, "find cloudlet")
	assert.Equal(t, appInst.Uri, reply.Fqdn)

	// Check GetAppInstList API - check sorted by distance
	ctx = dmecommon.PeerContext(context.Background(), "127.0.0.1", 123, span)
	reg = dme.RegisterClientRequest{
		OrgName: dmetest.Apps[1].Organization,
		AppName: dmetest.Apps[1].Name,
		AppVers: dmetest.Apps[1].Vers,
	}
	regReply, err = serv.RegisterClient(ctx, &reg)
	require.Nil(t, err)
	ckey, err = dmecommon.VerifyCookie(ctx, regReply.SessionCookie)
	require.Nil(t, err, "verify cookie")
	ctx = dmecommon.NewCookieContext(ctx, ckey)

	listReq := dme.AppInstListRequest{}
	listReq.GpsLocation = &dme.Loc{
		Latitude:  51,
		Longitude: 11,
	}
	listReq.SessionCookie = regReply.SessionCookie
	listReq.Limit = 4
	listReply, err := serv.GetAppInstList(ctx, &listReq)
	require.Nil(t, err)
	require.Equal(t, 4, len(listReply.Cloudlets))
	lastDist := float64(0)
	for _, cloc := range listReply.Cloudlets {
		require.Greater(t, cloc.Distance, lastDist)
		lastDist = cloc.Distance
	}

	// delete all data
	for _, app := range apps {
		dmecommon.RemoveApp(ctx, app)
	}
	assert.Equal(t, 0, len(tbl.Apps))
}

type dummyDmeApp struct {
	insts map[edgeproto.CloudletKey]struct{}
}

func checkAllData(t *testing.T, appInsts []*edgeproto.AppInst) {
	tbl := dmecommon.DmeAppTbl

	appsCheck := make(map[edgeproto.AppKey]*dummyDmeApp)
	for _, inst := range appInsts {
		app, found := appsCheck[inst.Key.AppKey]
		if !found {
			app = &dummyDmeApp{}
			app.insts = make(map[edgeproto.CloudletKey]struct{})
			appsCheck[inst.Key.AppKey] = app
		}
		app.insts[inst.Key.ClusterInstKey.CloudletKey] = struct{}{}
	}
	assert.Equal(t, len(appsCheck), len(tbl.Apps), "Number of apps")
	totalInstances := 0
	for k, app := range tbl.Apps {
		_, found := appsCheck[k]
		assert.True(t, found, "found app %s", k)
		if !found {
			continue
		}
		for cname := range app.Carriers {
			totalInstances += len(app.Carriers[cname].Insts)
		}
	}
	assert.Equal(t, totalInstances, len(appInsts), "Number of appInstances")
}
