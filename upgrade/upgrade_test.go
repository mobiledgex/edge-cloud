package upgrade

import (
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/assert"
)

var testCloudletKey = edgeproto.CloudletKey{
	OperatorKey: edgeproto.OperatorKey{
		Name: "testoperator",
	},
	Name: "testcloudlet",
}
var testCloudlet = edgeproto.Cloudlet{
	Key:           testCloudletKey,
	AccessUri:     "10.100.0.1",
	IpSupport:     edgeproto.IpSupport_IpSupportDynamic,
	NumDynamicIps: 100,
	Location: dme.Loc{
		Latitude:  100,
		Longitude: 100,
	},
}
var testClusterInstKey = edgeproto.ClusterInstKey{
	ClusterKey: edgeproto.ClusterKey{
		Name: "testcluster",
	},
	CloudletKey: testCloudletKey,
}
var appInstV0 = AppInstV0{
	Key: AppInstKeyV0{
		AppKey: edgeproto.AppKey{
			DeveloperKey: edgeproto.DeveloperKey{
				Name: "testdeveloper",
			},
			Name:    "testapp",
			Version: "1.0.0",
		},
		CloudletKey: testCloudletKey,
		Id:          1,
	},
	ClusterInstKey: testClusterInstKey,
	CloudletLoc:    testCloudlet.Location,
}

var appInstV1 = edgeproto.AppInst{
	Key: edgeproto.AppInstKey{
		AppKey: edgeproto.AppKey{
			DeveloperKey: edgeproto.DeveloperKey{
				Name: "testdeveloper",
			},
			Name:    "testapp",
			Version: "1.0.0",
		},
		ClusterInstKey: testClusterInstKey,
	},
	CloudletLoc: testCloudlet.Location,
	Version:     1,
}

func TestAppInstV0toV1Upgrade(t *testing.T) {
	testAppInst := appInstV0
	// test unsupported upgrade
	testAppInst.Version = 1
	res, err := UpgradeAppInstV0toV1(&testAppInst)
	assert.Nil(t, res, "upgrade of an incorrect version")
	assert.EqualError(t, err, ErrUpgradNotSupported.Error(), "incorrect error returned")
	// test supported upgrade
	testAppInst.Version = 0
	res, err = UpgradeAppInstV0toV1(&testAppInst)
	assert.Nil(t, err, "Unable to check correct version")
	assert.NotNil(t, res, "didn't get a new app back")
	assert.Equal(t, &appInstV1, res, "result doesn't match the expected appInst")

}
func TestAppInstV1toV0Downgrade(t *testing.T) {
	testAppInst := appInstV1
	// test unsupported downgrade
	testAppInst.Version = 2
	res, err := DowngradeAppInstV1toV0(&testAppInst)
	assert.Nil(t, res, "downgrade of an incorrect version")
	assert.EqualError(t, err, ErrUpgradNotSupported.Error(), "incorrect error returned")
	// test supported upgrade
	testAppInst.Version = 1
	res, err = DowngradeAppInstV1toV0(&testAppInst)
	assert.Nil(t, err, "Unable to check correct version")
	assert.NotNil(t, res, "didn't get a new app back")
	assert.Equal(t, &appInstV0, res, "result doesn't match the expected appInst")
}
