package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type CloudletVMPoolApi struct {
	sync  *Sync
	store edgeproto.CloudletVMPoolStore
	cache edgeproto.CloudletVMPoolCache
}

var cloudletVMPoolApi = CloudletVMPoolApi{}

func InitCloudletVMPoolApi(sync *Sync) {
	cloudletVMPoolApi.sync = sync
	cloudletVMPoolApi.store = edgeproto.NewCloudletVMPoolStore(sync.store)
	edgeproto.InitCloudletVMPoolCache(&cloudletVMPoolApi.cache)
	sync.RegisterCache(&cloudletVMPoolApi.cache)
}

func (s *CloudletVMPoolApi) AddCloudletVMPoolMember(ctx context.Context, in *edgeproto.CloudletVMPoolMember) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *CloudletVMPoolApi) RemoveCloudletVMPoolMember(ctx context.Context, in *edgeproto.CloudletVMPoolMember) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *CloudletVMPoolApi) CreateCloudletVMPool(ctx context.Context, in *edgeproto.CloudletVMPool) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *CloudletVMPoolApi) UpdateCloudletVMPool(ctx context.Context, in *edgeproto.CloudletVMPool) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *CloudletVMPoolApi) DeleteCloudletVMPool(ctx context.Context, in *edgeproto.CloudletVMPool) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *CloudletVMPoolApi) ShowCloudletVMPool(in *edgeproto.CloudletVMPool, cb edgeproto.CloudletVMPoolApi_ShowCloudletVMPoolServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletVMPool) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
