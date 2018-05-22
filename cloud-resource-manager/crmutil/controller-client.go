package crmutil

import (
	"context"
	"fmt"

	"github.com/bobbae/q"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func CloudletClient(api edgeproto.CloudletApiClient, srv *CloudResourceManagerServer) error {
	if len(srv.CloudResourceData.CloudResources) < 3 {
		return fmt.Errorf("At least 3 Cloud Resource items required.")
	}

	err := prepCloudletData(srv)
	ctx := context.TODO()

	srv.mux.Lock()
	defer srv.mux.Unlock()

	q.Q("call cloudlet api", len(srv.CloudResourceData.CloudResources))

	for i, cr := range srv.CloudResourceData.CloudResources {
		q.Q(cr)
		_, err = api.CreateCloudlet(ctx, &CloudletData[i])
		if err != nil {
			q.Q(err)
			break
		}
		q.Q("CreateCloudlet", CloudletData[i])
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
