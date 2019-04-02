package upgrade

import (
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var ErrUpgradNotSupported = errors.New("Unsupported upgrade path")

func UpgradeAppInstV0toV1(appInst *AppInstV0) (*edgeproto.AppInst, error) {
	if appInst.Version != 0 {
		log.DebugLog(log.DebugLevelUpgrade, "Already at a non zero version", "key", appInst.Key,
			"version", appInst.Version)
		return nil, ErrUpgradNotSupported
	}
	newAppInst := edgeproto.AppInst{}
	newAppInst.Fields = appInst.Fields
	newAppInst.Key.AppKey = appInst.Key.AppKey
	newAppInst.Key.ClusterInstKey = appInst.ClusterInstKey
	newAppInst.CloudletLoc = appInst.CloudletLoc
	newAppInst.Uri = appInst.Uri
	newAppInst.Liveness = appInst.Liveness
	newAppInst.MappedPorts = appInst.MappedPorts
	newAppInst.Flavor = appInst.Flavor
	newAppInst.State = appInst.State
	newAppInst.Errors = appInst.Errors
	newAppInst.CrmOverride = appInst.CrmOverride
	newAppInst.CreatedAt = appInst.CreatedAt
	newAppInst.Version = 1
	return &newAppInst, nil
}

func DowngradeAppInstV1toV0(appInst *edgeproto.AppInst) (*AppInstV0, error) {
	if appInst.Version != 1 {
		log.DebugLog(log.DebugLevelUpgrade, "Downgrading an incorrect version", "key", appInst.Key,
			"version", appInst.Version)
		return nil, ErrUpgradNotSupported
	}
	oldAppInst := AppInstV0{}
	oldAppInst.Fields = appInst.Fields
	oldAppInst.Key.AppKey = appInst.Key.AppKey
	oldAppInst.Key.CloudletKey = appInst.Key.ClusterInstKey.CloudletKey
	oldAppInst.Key.Id = 1 // Assume only a single appInst per cluster
	oldAppInst.CloudletLoc = appInst.CloudletLoc
	oldAppInst.Uri = appInst.Uri
	oldAppInst.ClusterInstKey = appInst.Key.ClusterInstKey
	oldAppInst.Liveness = appInst.Liveness
	oldAppInst.MappedPorts = appInst.MappedPorts
	oldAppInst.Flavor = appInst.Flavor
	oldAppInst.State = appInst.State
	oldAppInst.Errors = appInst.Errors
	oldAppInst.CrmOverride = appInst.CrmOverride
	oldAppInst.CreatedAt = appInst.CreatedAt
	oldAppInst.Version = 0
	return &oldAppInst, nil
}
