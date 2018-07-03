package crmutil

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type ControllerData struct {
	AppInstCache      edgeproto.AppInstCache
	CloudletCache     edgeproto.CloudletCache
	AppInstInfoCache  edgeproto.AppInstInfoCache
	CloudletInfoCache edgeproto.CloudletInfoCache
}

// NewControllerData creates a new instance to track data from the controller
func NewControllerData() *ControllerData {
	cd := &ControllerData{}
	edgeproto.InitAppInstCache(&cd.AppInstCache)
	edgeproto.InitCloudletCache(&cd.CloudletCache)
	edgeproto.InitAppInstInfoCache(&cd.AppInstInfoCache)
	edgeproto.InitCloudletInfoCache(&cd.CloudletInfoCache)
	return cd
}
