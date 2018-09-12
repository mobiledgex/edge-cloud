package crmutil

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

//CloudResourceData contains resources
type CloudResourceData struct {
	CloudResources []*edgeproto.CloudResource
}

var crdb = CloudResourceData{
	CloudResources: crs,
}

var crs = []*edgeproto.CloudResource{
	&edgeproto.CloudResource{
		Name:     "Cloudlet-1",
		Category: edgeproto.CloudResourceCategory_Kubernetes,
		Active:   false,
		Id:       0,
	},
	&edgeproto.CloudResource{
		Name:     "Cloudlet-2",
		Category: edgeproto.CloudResourceCategory_Kubernetes,
		Active:   false,
		Id:       1,
	},
	&edgeproto.CloudResource{
		Name:     "Cloudlet-3",
		Category: edgeproto.CloudResourceCategory_Kubernetes,
		Active:   false,
		Id:       2,
	},
}

var resourceID int32

// CloudResourceManagerServer describes Cloud Resource Manager Server instance container
type CloudResourceManagerServer struct {
	mux               sync.Mutex
	CloudResourceData CloudResourceData
	ControllerData    *ControllerData
}

// NewCloudResourceManagerServer Creates new Cloud Resource Manager Service instance
func NewCloudResourceManagerServer(cd *ControllerData) (*CloudResourceManagerServer, error) {
	return &CloudResourceManagerServer{CloudResourceData: crdb, ControllerData: cd}, nil
}

//ListCloudResource lists resources
func (server *CloudResourceManagerServer) ListCloudResource(cr *edgeproto.CloudResource, cb edgeproto.CloudResourceManager_ListCloudResourceServer) error {
	var err error

	server.mux.Lock()
	defer server.mux.Unlock()

	for _, obj := range server.CloudResourceData.CloudResources {
		if cr.Category != 0 && cr.Category != obj.Category {
			continue
		}

		err = cb.Send(obj)
		if err != nil {
			log.Printf("Can't strearm out resource, %v", err)
			break
		}
	}

	return err
}

// AddCloudResource adds new resource
func (server *CloudResourceManagerServer) AddCloudResource(ctx context.Context, cr *edgeproto.CloudResource) (*edgeproto.Result, error) {
	server.mux.Lock()
	defer server.mux.Unlock()

	cr.Id = resourceID
	resourceID = resourceID + 1

	server.CloudResourceData.CloudResources = append(server.CloudResourceData.CloudResources, cr)
	cloudlet := edgeproto.Cloudlet{}
	found := server.ControllerData.CloudletCache.Get(cr.CloudletKey, &cloudlet)
	if !found {
		err := fmt.Errorf("cloudlet not found %v", cr)
		errstr := fmt.Sprintf("error %v", err)
		return &edgeproto.Result{Message: errstr}, err
	}

	return &edgeproto.Result{}, nil
}

//DeleteCloudResource removes a resource
func (server *CloudResourceManagerServer) DeleteCloudResource(ctx context.Context, cr *edgeproto.CloudResource) (*edgeproto.Result, error) {

	server.mux.Lock()
	defer server.mux.Unlock()

	found := false
	foundIndex := -1
	for i, obj := range server.CloudResourceData.CloudResources {
		if cr.Name == obj.Name {
			found = true
			foundIndex = i
		}
	}
	if found {
		server.CloudResourceData.CloudResources = append(server.CloudResourceData.CloudResources[:foundIndex], server.CloudResourceData.CloudResources[foundIndex+1:]...)
		return &edgeproto.Result{}, nil
	}

	return nil, fmt.Errorf("Resource not found")
}

//DeployApplication runs app
func (server *CloudResourceManagerServer) DeployApplication(ctx context.Context, app *edgeproto.EdgeCloudApplication) (*edgeproto.Result, error) {
	if err := RunApp(app); err != nil {
		return nil, err
	}

	appInst := edgeproto.AppInst{}
	found := server.ControllerData.AppInstCache.Get(app.Apps[0].AppInstKey, &appInst)
	if !found {
		// controller has no such app inst, should fail
		err := fmt.Errorf("app not found %v", app)
		errstr := fmt.Sprintf("error %v", err)
		return &edgeproto.Result{Message: errstr}, err
	}

	return &edgeproto.Result{}, nil
}

// DeleteApplication removes app
func (server *CloudResourceManagerServer) DeleteApplication(ctx context.Context, app *edgeproto.EdgeCloudApplication) (*edgeproto.Result, error) {
	if err := KillApp(app); err != nil {
		return nil, err
	}
	return &edgeproto.Result{}, nil
}
