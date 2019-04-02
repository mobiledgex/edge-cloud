package upgrade

import "github.com/mobiledgex/edge-cloud/edgeproto"

func UpgradeAppInstV0toV1(appInst *AppInstV0) (*edgeproto.AppInst, error) {
	newAppInst := edgeproto.AppInst{}
	return &newAppInst, nil
}

func DowngradeAppInstV1toV0(appInst *edgeproto.AppInst) (*AppInstV0, error) {
	oldAppInst := AppInstV0{}
	return &oldAppInst, nil
}
