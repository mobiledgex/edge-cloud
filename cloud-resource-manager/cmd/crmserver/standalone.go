package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type standaloneServer struct {
	data *crmutil.ControllerData
}

func (s *standaloneServer) CreateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) error {
	s.data.AppInstCache.Update(in, 0)
	return nil
}

func (s *standaloneServer) DeleteAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	s.data.AppInstCache.Delete(in, 0)
	return nil
}

func (s *standaloneServer) UpdateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_UpdateAppInstServer) error {
	s.data.AppInstCache.Update(in, 0)
	return nil
}

func (s *standaloneServer) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	err := s.data.AppInstCache.Show(in, func(obj *edgeproto.AppInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}



func (s *standaloneServer) CreateClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_CreateClusterInstServer) error {
	s.data.ClusterInstCache.Update(in, 0)
	return nil
}

func (s *standaloneServer) DeleteClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	s.data.ClusterInstCache.Delete(in, 0)
	return nil
}

func (s *standaloneServer) UpdateClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_UpdateClusterInstServer) error {
	s.data.ClusterInstCache.Update(in, 0)
	return nil
}

func (s *standaloneServer) ShowClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_ShowClusterInstServer) error {
	err := s.data.ClusterInstCache.Show(in, func(obj *edgeproto.ClusterInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *standaloneServer) CreateCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_CreateCloudletServer) error {
	s.data.CloudletCache.Update(in, 0)
	return nil
}

func (s *standaloneServer) DeleteCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
	s.data.CloudletCache.Delete(in, 0)
	return nil
}

func (s *standaloneServer) UpdateCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_UpdateCloudletServer) error {
	s.data.CloudletCache.Update(in, 0)
	return nil
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

func (s *standaloneServer) InjectCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	s.data.CloudletInfoCache.Update(in, 0)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) EvictCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	s.data.CloudletInfoCache.Delete(in, 0)
	return &edgeproto.Result{}, nil
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

func (s *standaloneServer) ShowClusterInstInfo(in *edgeproto.ClusterInstInfo, cb edgeproto.ClusterInstInfoApi_ShowClusterInstInfoServer) error {
	err := s.data.ClusterInstInfoCache.Show(in, func(obj *edgeproto.ClusterInstInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
