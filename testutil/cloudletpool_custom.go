package testutil

import (
	"context"
	fmt "fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) AddCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	cache := &s.CloudletPoolCache

	cache.Mux.Lock()
	defer cache.Mux.Unlock()
	data, found := cache.Objs[in.Key]
	if !found {
		return &edgeproto.Result{}, in.Key.NotFoundError()
	}
	for ii, _ := range data.Obj.Cloudlets {
		if data.Obj.Cloudlets[ii] == in.CloudletName {
			return &edgeproto.Result{}, fmt.Errorf("Already exists")
		}
	}
	data.Obj.Cloudlets = append(data.Obj.Cloudlets, in.CloudletName)

	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	cache := &s.CloudletPoolCache

	cache.Mux.Lock()
	defer cache.Mux.Unlock()
	data, found := cache.Objs[in.Key]
	if !found {
		return &edgeproto.Result{}, in.Key.NotFoundError()
	}
	for ii, cloudlet := range data.Obj.Cloudlets {
		if cloudlet == in.CloudletName {
			data.Obj.Cloudlets = append(data.Obj.Cloudlets[:ii], data.Obj.Cloudlets[ii+1:]...)
			break
		}
	}

	return &edgeproto.Result{}, nil
}
