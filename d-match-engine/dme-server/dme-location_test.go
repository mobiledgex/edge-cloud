package main

import (
	"testing"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestVerifyLoc(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	span := log.SpanFromContext(ctx)

	dmecommon.SetupMatchEngine()
	InitAppInstClients()
	operatorApiGw, _ = initOperator(ctx, "standalone")
	setupJwks()

	// add all data
	for _, app := range dmetest.GenerateApps() {
		dmecommon.AddApp(ctx, app)
	}
	for _, inst := range dmetest.GenerateAppInsts() {
		dmecommon.AddAppInst(ctx, inst)
	}
	serv := server{}
	// test verify location
	for ii, rr := range dmetest.VerifyLocData {
		ctx := dmecommon.PeerContext(context.Background(), "127.0.0.1", 123, span)

		regReply, err := serv.RegisterClient(ctx, &rr.Reg)
		assert.Nil(t, err, "register client")

		// Since we're directly calling functions, we end up
		// bypassing the interceptor which sets up the cookie key.
		// So set it on the context manually.
		ckey, err := dmecommon.VerifyCookie(ctx, regReply.SessionCookie)
		require.Nil(t, err, "verify cookie")
		ctx = dmecommon.NewCookieContext(ctx, ckey)

		reply, err := serv.VerifyLocation(ctx, &rr.Req)
		if err != nil {
			assert.Contains(t, err.Error(), rr.Error, "VerifyLocData[%d]", ii)
		} else {
			assert.Equal(t, &rr.Reply, reply, "VerifyLocData[%d]", ii)
		}
	}
}

func setupJwks() {
	// setup fake JWT key
	config := vault.NewConfig("foo", vault.NewAppRoleAuth("roleID", "secretID"))
	dmecommon.Jwks.Init(config, "local", "dme")
	dmecommon.Jwks.Meta.CurrentVersion = 1
	dmecommon.Jwks.Keys[1] = &vault.JWK{
		Secret:  "12345",
		Refresh: "1s",
	}
}
