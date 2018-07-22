package crmutil

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

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

var resourceID int32 = 0

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

// List Cloud Resource
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

// Add Cloud Resource
func (server *CloudResourceManagerServer) AddCloudResource(ctx context.Context, cr *edgeproto.CloudResource) (*edgeproto.Result, error) {
	server.mux.Lock()
	defer server.mux.Unlock()

	cr.Id = resourceID
	resourceID = resourceID + 1

	server.CloudResourceData.CloudResources = append(server.CloudResourceData.CloudResources, cr)
	cloudlet := edgeproto.Cloudlet{}
	found := server.ControllerData.CloudletCache.Get(cr.CloudletKey, &cloudlet)
	if !found {
		// controller has no such cloudlet, should fail
	}

	return &edgeproto.Result{}, nil
}

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

func (server *CloudResourceManagerServer) DeployApplication(ctx context.Context, app *edgeproto.EdgeCloudApplication) (*edgeproto.Result, error) {
	if err := RunApp(app); err != nil {
		return nil, err
	}

	appInst := edgeproto.AppInst{}
	found := server.ControllerData.AppInstCache.Get(app.Apps[0].AppInstKey, &appInst)
	if !found {
		// controller has no such app inst, should fail
	}

	return &edgeproto.Result{}, nil
}

func (server *CloudResourceManagerServer) DeleteApplication(ctx context.Context, app *edgeproto.EdgeCloudApplication) (*edgeproto.Result, error) {
	if err := KillApp(app); err != nil {
		return nil, err
	}
	return &edgeproto.Result{}, nil
}
