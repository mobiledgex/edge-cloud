package main

import (
	"fmt"
	"testing"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestAddRemove(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq | log.DebugLevelDmedb)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	span := log.SpanFromContext(ctx)

	eehandler, err := initEdgeEventsPlugin(ctx, "standalone")
	require.Nil(t, err, "init edge events plugin")
	dmecommon.SetupMatchEngine(eehandler)
	dmecommon.InitAppInstClients()
	defer dmecommon.StopAppInstClients()

	setupJwks()
	apps := dmetest.GenerateApps()
	appInsts := dmetest.GenerateAppInsts()
	cloudlets := dmetest.GenerateCloudlets()

	tbl := dmecommon.DmeAppTbl

	// Add cloudlets first as we check the state via cloudlets
	for _, cloudlet := range cloudlets {
		dmecommon.SetInstStateFromCloudletInfo(ctx, cloudlet)
	}
	require.Equal(t, len(tbl.Cloudlets), 0, "without cloudlet object, cloudletInfo is not considered")
	for _, cloudlet := range cloudlets {
		dmecommon.SetInstStateFromCloudlet(ctx, &edgeproto.Cloudlet{Key: cloudlet.Key})
		dmecommon.SetInstStateFromCloudletInfo(ctx, cloudlet)
	}
	require.Equal(t, len(tbl.Cloudlets), len(cloudlets), "cloudlet object exists")

	// add alliance orgs
	cloudletShared := edgeproto.Cloudlet{
		Key:          cloudlets[1].Key,
		AllianceOrgs: []string{"DMUUS"},
	}
	dmecommon.SetInstStateFromCloudlet(ctx, &cloudletShared)

	// add all data, check that number of instances matches
	for _, inst := range apps {
		dmecommon.AddApp(ctx, inst)
	}
	for _, inst := range appInsts {
		dmecommon.AddAppInst(ctx, inst)
	}
	checkAllData(t, appInsts)
	// only one cloudlet with one alliance org, and since all apps are
	// deployed to each cloudlet, that means one set of appinsts are
	// added to the alliance count for cloudlets[1].
	checkAllianceInsts(t, len(apps))

	// re-add data, counts should remain unchanged
	for _, inst := range appInsts {
		dmecommon.AddAppInst(ctx, inst)
	}
	checkAllData(t, appInsts)
	checkAllianceInsts(t, len(apps))

	// delete one data, check new counts
	dmecommon.RemoveAppInst(ctx, appInsts[0])
	remaining := appInsts[1:]
	checkAllData(t, remaining)
	checkAllianceInsts(t, len(apps))
	serv := server{}

	// test findCloudlet
	runFindCloudlet(t, dmetest.FindCloudletData, span, &serv)
	runFindCloudlet(t, dmetest.FindCloudletAllianceOrg, span, &serv)
	runGetAppInstList(t, dmetest.GetAppInstListAllianceOrg, span, &serv)

	// update cloudlet alliance orgs to remove them
	cloudletNotShared := cloudletShared
	cloudletNotShared.AllianceOrgs = []string{}
	dmecommon.SetInstStateFromCloudlet(ctx, &cloudletNotShared)
	// run checks
	checkAllianceInsts(t, 0)
	runFindCloudlet(t, dmetest.FindCloudletData, span, &serv)
	runFindCloudlet(t, dmetest.FindCloudletNoAllianceOrg, span, &serv)
	runGetAppInstList(t, dmetest.GetAppInstListNoAllianceOrg, span, &serv)

	// add back alliance orgs
	dmecommon.SetInstStateFromCloudlet(ctx, &cloudletShared)
	// run checks
	checkAllianceInsts(t, len(apps))
	runFindCloudlet(t, dmetest.FindCloudletData, span, &serv)
	runFindCloudlet(t, dmetest.FindCloudletAllianceOrg, span, &serv)
	runGetAppInstList(t, dmetest.GetAppInstListAllianceOrg, span, &serv)

	// test findCloudlet HA. Repeat the FindCloudlet 100 times and
	// make sure we get results for both cloudlets
	for ii, rr := range dmetest.FindCloudletHAData {
		ctx := dmecommon.PeerContext(context.Background(), "127.0.0.1", 123, span)
		numFindsCloudlet1 := 0
		numFindsCloudlet2 := 0
		maxAttempts := 100
		minExpectedEachCloudlet := 35
		regReply, err := serv.RegisterClient(ctx, &rr.Reg)
		assert.Nil(t, err, "register client")
		ckey, err := dmecommon.VerifyCookie(ctx, regReply.SessionCookie)
		assert.Nil(t, err, "verify cookie")
		ctx = dmecommon.NewCookieContext(ctx, ckey)
		// Make sure we get the statsKey value filled in
		call := dmecommon.ApiStatCall{}
		ctx = context.WithValue(ctx, dmecommon.StatKeyContextKey, &call.Key)

		for attempt := 0; attempt < maxAttempts; attempt++ {

			reply, err := serv.FindCloudlet(ctx, &rr.Req)
			assert.Nil(t, err, "find cloudlet")
			assert.Equal(t, rr.Reply.Status, reply.Status, "findCloudletHAData[%d]", ii)
			if reply.Status == dme.FindCloudletReply_FIND_FOUND {
				if reply.Fqdn == rr.Reply.Fqdn {
					numFindsCloudlet1++
				} else if reply.Fqdn == rr.ReplyAlternate.Fqdn {
					numFindsCloudlet2++
				}
				// carrier is the same either way
				assert.Equal(t, rr.ReplyCarrier,
					call.Key.CloudletFound.Organization, "findCloudletHAData[%d]", ii)
			}
		}
		// we expect at least 35% of all replies to be for each cloudlet, with confidence of 99.8%
		assert.GreaterOrEqual(t, numFindsCloudlet1, minExpectedEachCloudlet)
		assert.GreaterOrEqual(t, numFindsCloudlet2, minExpectedEachCloudlet)
		// total for both should match attempts
		assert.Equal(t, maxAttempts, numFindsCloudlet1+numFindsCloudlet2)
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
	cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_UNKNOWN
	dmecommon.SetInstStateFromCloudletInfo(ctx, cloudletInfo)
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
	cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_READY
	dmecommon.SetInstStateFromCloudletInfo(ctx, cloudletInfo)
	reply, err = serv.FindCloudlet(ctx, &dmetest.DisabledCloudletRR.Req)
	assert.Nil(t, err, "find cloudlet")
	assert.Equal(t, dmetest.FindCloudletData[3].Reply.Status, reply.Status)
	assert.Equal(t, dmetest.FindCloudletData[3].Reply.Fqdn, reply.Fqdn)

	// Change the health check status of the appInst and get check the results
	appInst := dmetest.MakeAppInst(&dmetest.Apps[0], &dmetest.Cloudlets[2])
	appInst.HealthCheck = dme.HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE
	dmecommon.AddAppInst(ctx, appInst)
	reply, err = serv.FindCloudlet(ctx, &dmetest.DisabledCloudletRR.Req)
	assert.Nil(t, err, "find cloudlet")
	assert.Equal(t, dmetest.DisabledCloudletRR.Reply.Status, reply.Status)
	assert.Equal(t, dmetest.DisabledCloudletRR.Reply.Fqdn, reply.Fqdn)
	// reset and check the one that we get is returned
	appInst.HealthCheck = dme.HealthCheck_HEALTH_CHECK_OK
	dmecommon.AddAppInst(ctx, appInst)
	reply, err = serv.FindCloudlet(ctx, &dmetest.DisabledCloudletRR.Req)
	assert.Nil(t, err, "find cloudlet")
	assert.Equal(t, appInst.Uri, reply.Fqdn)

	// Check GetAppInstList API - check sorted by distance
	runGetAppInstList(t, dmetest.GetAppInstListData, span, &serv)

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

func checkAllianceInsts(t *testing.T, expectedCount int) {
	tbl := dmecommon.DmeAppTbl
	total := 0
	for _, app := range tbl.Apps {
		for cname := range app.Carriers {
			total += len(app.Carriers[cname].AllianceInsts)
		}
	}
	require.Equal(t, expectedCount, total, "Number of alliance appInstances")
}

func runFindCloudlet(t *testing.T, rrs []dmetest.FindCloudletRR, span opentracing.Span, serv *server) {
	for ii, rr := range rrs {
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
		call := dmecommon.ApiStatCall{}
		ctx = context.WithValue(ctx, dmecommon.StatKeyContextKey, &call.Key)

		reply, err := serv.FindCloudlet(ctx, &rr.Req)
		assert.Nil(t, err, "find cloudlet")
		assert.Equal(t, rr.Reply.Status, reply.Status, "findCloudletData[%d]", ii)
		if reply.Status == dme.FindCloudletReply_FIND_FOUND {
			require.Equal(t, rr.Reply.Fqdn, reply.Fqdn,
				"findCloudletData[%d]", ii)
			// Check the filled in cloudlet details
			require.Equal(t, rr.ReplyCarrier,
				call.Key.CloudletFound.Organization, "findCloudletData[%d]", ii)
			require.Equal(t, rr.ReplyCloudlet,
				call.Key.CloudletFound.Name, "findCloudletData[%d]", ii)
		}
	}
}

func runGetAppInstList(t *testing.T, rrs []dmetest.GetAppInstListRR, span opentracing.Span, serv *server) {
	for ii, rr := range rrs {
		ctx := dmecommon.PeerContext(context.Background(), "127.0.0.1", 123, span)
		info := fmt.Sprintf("[%d]", ii)

		regReply, err := serv.RegisterClient(ctx, &rr.Reg)
		require.Nil(t, err, info)
		ckey, err := dmecommon.VerifyCookie(ctx, regReply.SessionCookie)
		require.Nil(t, err, info)
		// set session cookie key directly on context since we're bypassing
		// interceptors
		ctx = dmecommon.NewCookieContext(ctx, ckey)

		reply, err := serv.GetAppInstList(ctx, &rr.Req)
		require.Nil(t, err, info)
		require.Equal(t, rr.Reply.Status, reply.Status, info)
		require.Equal(t, len(rr.Reply.Cloudlets), len(reply.Cloudlets), info)
		var lastDistance float64
		for jj, clExp := range rr.Reply.Cloudlets {
			clAct := reply.Cloudlets[jj]
			info2 := fmt.Sprintf("[%d][%d]", ii, jj)

			require.NotNil(t, clAct, info2)
			require.Equal(t, clExp.CarrierName, clAct.CarrierName, info2)
			require.Equal(t, clExp.CloudletName, clAct.CloudletName, info2)
			require.NotNil(t, clAct.GpsLocation, info2)
			require.Equal(t, clExp.GpsLocation.Latitude, clAct.GpsLocation.Latitude, info2)
			require.Equal(t, clExp.GpsLocation.Longitude, clAct.GpsLocation.Longitude, info2)
			require.GreaterOrEqual(t, clAct.Distance, lastDistance, info2)
			lastDistance = clAct.Distance
		}
	}
}
