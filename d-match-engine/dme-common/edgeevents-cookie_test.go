package dmecommon

import (
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

func TestEdgeEventsCookie(t *testing.T) {
	// Init context
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	setupJwks()
	// Init variables to be put into cookie
	dmeAppInst := &DmeAppInst{
		virtualClusterInstKey: edgeproto.VirtualClusterInstKey{
			Organization: "testcluster-org",
			ClusterKey: edgeproto.ClusterKey{
				Name: "testcluster",
			},
			CloudletKey: edgeproto.CloudletKey{
				Organization: "testcloudlet-org",
				Name:         "testcloudlet",
			},
		},
	}
	location := dme.Loc{
		Latitude:           10.0,
		Longitude:          11.0,
		HorizontalAccuracy: 123.4,
		VerticalAccuracy:   321.03,
		Altitude:           1.453423,
		Course:             45.3,
		Speed:              3.453,
		Timestamp: &dme.Timestamp{
			Seconds: time.Now().Unix(),
		},
	}
	// Test create and verify cookie with all the values
	key := CreateEdgeEventsCookieKey(dmeAppInst, location)
	cookieFromKey, err := GenerateEdgeEventsCookie(key, ctx, 10*time.Minute)
	assert.Nil(t, err)
	assert.NotEqual(t, "", cookieFromKey)
	keyFromCookie, err := VerifyEdgeEventsCookie(ctx, cookieFromKey)
	assert.Nil(t, err)
	assert.NotNil(t, keyFromCookie)
	assert.Equal(t, *key, *keyFromCookie)
}

func setupJwks() {
	// setup fake JWT key
	config := vault.NewConfig("foo", vault.NewAppRoleAuth("roleID", "secretID"))
	Jwks.Init(config, "local", "dme")
	Jwks.Meta.CurrentVersion = 1
	Jwks.Keys[1] = &vault.JWK{
		Secret:  "12345",
		Refresh: "1s",
	}
}
