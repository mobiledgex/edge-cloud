package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type standaloneServer struct {
	data *crmutil.ControllerData
}

func (s *standaloneServer) CreateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	s.data.AppInstCache.Update(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) DeleteAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	s.data.AppInstCache.Delete(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) UpdateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	s.data.AppInstCache.Update(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	err := s.data.AppInstCache.Show(in, func(obj *edgeproto.AppInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *standaloneServer) CreateClusterInst(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	s.data.ClusterInstCache.Update(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) DeleteClusterInst(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	s.data.ClusterInstCache.Delete(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) UpdateClusterInst(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	s.data.ClusterInstCache.Update(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) ShowClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_ShowClusterInstServer) error {
	err := s.data.ClusterInstCache.Show(in, func(obj *edgeproto.ClusterInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *standaloneServer) CreateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	s.data.CloudletCache.Update(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) DeleteCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	s.data.CloudletCache.Delete(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) UpdateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	s.data.CloudletCache.Update(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) ShowCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_ShowCloudletServer) error {
	err := s.data.CloudletCache.Show(in, func(obj *edgeproto.Cloudlet) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *standaloneServer) CreateFlavor(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {
	s.data.FlavorCache.Update(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) DeleteFlavor(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {
	s.data.FlavorCache.Delete(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) UpdateFlavor(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {
	s.data.FlavorCache.Update(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) ShowFlavor(in *edgeproto.Flavor, cb edgeproto.FlavorApi_ShowFlavorServer) error {
	err := s.data.FlavorCache.Show(in, func(obj *edgeproto.Flavor) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *standaloneServer) ShowCloudletInfo(in *edgeproto.CloudletInfo, cb edgeproto.CloudletInfoApi_ShowCloudletInfoServer) error {
	err := s.data.CloudletInfoCache.Show(in, func(obj *edgeproto.CloudletInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *standaloneServer) ShowAppInstInfo(in *edgeproto.AppInstInfo, cb edgeproto.AppInstInfoApi_ShowAppInstInfoServer) error {
	err := s.data.AppInstInfoCache.Show(in, func(obj *edgeproto.AppInstInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
