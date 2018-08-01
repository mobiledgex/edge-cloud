package crmutil

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	log "github.com/sirupsen/logrus"
)

//CloudletClient calls CreateCloudlet
func CloudletClient(api edgeproto.CloudletApiClient, srv *CloudResourceManagerServer) error {
	if len(srv.CloudResourceData.CloudResources) < 3 {
		return fmt.Errorf("at least 3 Cloud Resource items required")
	}

	err := prepCloudletData(srv)
	ctx := context.TODO()

	srv.mux.Lock()
	defer srv.mux.Unlock()

	for i, cr := range srv.CloudResourceData.CloudResources {
		_, err = api.CreateCloudlet(ctx, &CloudletData[i])
		if err != nil {
			log.Errorf("error calling CreateCloudlet, %v %v", cr, err)
			break
		}
	}

	return err
}

func prepCloudletData(srv *CloudResourceManagerServer) error {
	srv.mux.Lock()
	defer srv.mux.Unlock()

	if len(CloudletData) < len(srv.CloudResourceData.CloudResources) {
		return fmt.Errorf("insufficient cloudlet test data")
	}

	for i, cr := range srv.CloudResourceData.CloudResources {
		cr.CloudletKey = &CloudletData[i].Key
	}

	return nil
}
