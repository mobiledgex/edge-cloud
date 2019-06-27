package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type PlatformApi struct {
	sync  *Sync
	store edgeproto.PlatformStore
	cache edgeproto.PlatformCache
}

const (
	PlatformInitTimeout = 5 * time.Minute
)

var platformApi = PlatformApi{}
var cloudletPlatform pf.Platform

func InitPlatformApi(sync *Sync) {
	platformApi.sync = sync
	platformApi.store = edgeproto.NewPlatformStore(sync.store)
	edgeproto.InitPlatformCache(&platformApi.cache)
	sync.RegisterCache(&platformApi.cache)
}

func (s *PlatformApi) CreatePlatform(ctx context.Context, in *edgeproto.Platform) (*edgeproto.Result, error) {
	if err := in.Validate(edgeproto.PlatformAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}

	parts := strings.Split(in.RegistryPath, "/")
	// Append default registry address for internal image paths
	if len(parts) < 2 || !strings.Contains(parts[0], ".") {
		return &edgeproto.Result{}, fmt.Errorf("registrypath should be full registry URL: <domain-name>/<registry-path>")
	}
	parts = strings.Split(in.RegistryPath, ":")
	if len(parts) > 1 {
		return &edgeproto.Result{}, fmt.Errorf("registrypath should not have image tag")

	}

	// Fetch Controller Image Tag from /version.txt
	// Platform image tag should be same as controller image tag
	platform_version, err := ioutil.ReadFile("/version.txt")
	if err != nil {
		return &edgeproto.Result{}, fmt.Errorf("unable to fetch controller image tag: %v", err)
	}

	platform_registry_path := in.RegistryPath + ":" + string(platform_version)
	if !*testMode {
		err = cloudcommon.ValidateDockerRegistryPath(platform_registry_path, *vaultAddr)
		if err != nil {
			return &edgeproto.Result{}, err
		}

		err = cloudcommon.ValidateVMRegistryPath(in.ImagePath, *vaultAddr)
		if err != nil {
			return &edgeproto.Result{}, err
		}
	}
	in.RegistryPath = platform_registry_path

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyExists
		}
		if !flavorApi.store.STMGet(stm, &in.Flavor, nil) {
			return fmt.Errorf("Flavor %s not found", in.Flavor.Name)
		}

		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}

	// Load platform implementation
	cloudletPlatform, err = pfutils.GetPlatform(in.PlatformType.String())
	if err != nil {
		return &edgeproto.Result{}, err
	}

	return s.store.Create(in, s.sync.syncWait)
}

func (s *PlatformApi) UpdatePlatform(ctx context.Context, in *edgeproto.Platform) (*edgeproto.Result, error) {
	// Unsupported for now
	return &edgeproto.Result{}, errors.New("Update platform not supported")
	//return s.store.Update(in, s.sync.syncWait)
}

func (s *PlatformApi) DeletePlatform(ctx context.Context, in *edgeproto.Platform) (*edgeproto.Result, error) {
	if cloudletApi.UsesPlatform(&in.Key) {
		return &edgeproto.Result{}, errors.New("Platform in use by Cloudlet")
	}
	cloudletPlatform = nil

	return s.store.Delete(in, s.sync.syncWait)
}

func (s *PlatformApi) ShowPlatform(in *edgeproto.Platform, cb edgeproto.PlatformApi_ShowPlatformServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Platform) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
