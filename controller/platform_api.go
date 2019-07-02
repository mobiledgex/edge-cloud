package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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

var (
	platformApi       = PlatformApi{}
	cloudletPlatforms = make(map[edgeproto.PlatformType]pf.Platform)
)

func InitPlatformApi(sync *Sync) {
	platformApi.sync = sync
	platformApi.store = edgeproto.NewPlatformStore(sync.store)
	edgeproto.InitPlatformCache(&platformApi.cache)
	sync.RegisterCache(&platformApi.cache)
}

func (s *PlatformApi) CreatePlatform(in *edgeproto.Platform, cb edgeproto.PlatformApi_CreatePlatformServer) error {
	var err error

	if err = in.Validate(edgeproto.PlatformAllFieldsMap); err != nil {
		return err
	}

	if !*testMode {
		if in.RegistryPath != "" {
			// Fetch Controller Image Tag from /version.txt
			// Platform image tag should be same as controller image tag
			platform_version, err := ioutil.ReadFile("/version.txt")
			if err != nil {
				return fmt.Errorf("unable to fetch controller image tag: %v", err)
			}
			tag := strings.TrimSpace(string(platform_version))
			platform_registry_path := in.RegistryPath + ":" + tag
			err = cloudcommon.ValidateDockerRegistryPath(platform_registry_path, *vaultAddr)
			if err != nil {
				return err
			}
			in.PlatformTag = tag
		}
		if in.ImagePath != "" {
			err = cloudcommon.ValidateVMRegistryPath(in.ImagePath, *vaultAddr)
			if err != nil {
				return err
			}
		}

		// Vault controller level credentials are required to access
		// registry credentials
		roleId := os.Getenv("VAULT_ROLE_ID")
		if roleId == "" {
			return fmt.Errorf("Env variable VAULT_ROLE_ID not set")
		}
		secretId := os.Getenv("VAULT_SECRET_ID")
		if secretId == "" {
			return fmt.Errorf("Env variable VAULT_SECRET_ID not set")
		}

		// Vault CRM level credentials are required to access
		// instantiate crmserver
		crmRoleId := os.Getenv("VAULT_CRM_ROLE_ID")
		if crmRoleId == "" {
			return fmt.Errorf("Env variable VAULT_CRM_ROLE_ID not set")
		}
		crmSecretId := os.Getenv("VAULT_CRM_SECRET_ID")
		if crmSecretId == "" {
			return fmt.Errorf("Env variable VAULT_CRM_SECRET_ID not set")
		}
		in.CrmRoleId = crmRoleId
		in.CrmSecretId = crmSecretId
	}

	in.TlsCertFile = *tlsCertFile
	in.VaultAddr = *vaultAddr

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !flavorApi.store.STMGet(stm, in.Flavor, nil) {
			return fmt.Errorf("Flavor %s not found", in.Flavor.Name)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Load platform implementation
	cloudletPlatform, ok := cloudletPlatforms[in.PlatformType]
	if !ok {
		cloudletPlatform, err = pfutils.GetPlatform(in.PlatformType.String())
		if err != nil {
			return err
		}
		cloudletPlatforms[in.PlatformType] = cloudletPlatform
	}

	updatePlatformCallback := func(updateType edgeproto.CacheUpdateType, value string) {
		switch updateType {
		case edgeproto.UpdateTask:
			log.DebugLog(log.DebugLevelApi, "SetStatusTask", "key", in.Key, "taskName", value)
			in.Status.SetTask(value)
			cb.Send(&edgeproto.Result{Message: in.Status.ToString()})
		case edgeproto.UpdateStep:
			log.DebugLog(log.DebugLevelApi, "SetStatusStep", "key", in.Key, "stepName", value)
			in.Status.SetStep(value)
			cb.Send(&edgeproto.Result{Message: in.Status.ToString()})
		}
	}

	in.State = edgeproto.TrackedState_CREATING
	_, err = s.store.Create(in, s.sync.syncWait)
	if err != nil {
		return err
	}

	// Create platform
	err = cloudletPlatform.CreatePlatform(in, updatePlatformCallback)
	if err != nil {
		in.State = edgeproto.TrackedState_CREATE_ERROR
		in.Errors = append(in.Errors, err.Error())
		_, err = s.store.Delete(in, s.sync.syncWait)
		return err
	}

	in.State = edgeproto.TrackedState_READY
	s.store.Put(in, s.sync.syncWait)
	return nil
}

func (s *PlatformApi) UpdatePlatform(in *edgeproto.Platform, cb edgeproto.PlatformApi_UpdatePlatformServer) error {
	// Unsupported for now
	return errors.New("Update platform not supported")
	//return s.store.Update(in, s.sync.syncWait)
}

func (s *PlatformApi) DeletePlatform(in *edgeproto.Platform, cb edgeproto.PlatformApi_DeletePlatformServer) error {
	if cloudletApi.UsesPlatform(&in.Key) {
		return errors.New("Platform in use by Cloudlet")
	}

	// Set state to prevent other apps from being created on ClusterInst
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return objstore.ErrKVStoreKeyNotFound
		}
		if in.State != edgeproto.TrackedState_READY && in.State != edgeproto.TrackedState_CREATE_ERROR {
			if in.State == edgeproto.TrackedState_DELETE_ERROR {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous delete failed, %v", in.Errors)})
				cb.Send(&edgeproto.Result{Message: "Use CreatePlatform to rebuild, and try again"})
			}
			return errors.New("Platform busy, cannot delete")
		}
		in.State = edgeproto.TrackedState_DELETE_PREPARE
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}

	cloudletPlatform, ok := cloudletPlatforms[in.PlatformType]
	if !ok {
		return fmt.Errorf("Platform plugin %s not found", in.PlatformType.String())
	}

	err = cloudletPlatform.DeletePlatform(in)
	if err != nil {
		return err
	}

	_, err = s.store.Delete(in, s.sync.syncWait)
	return err
}

func (s *PlatformApi) ShowPlatform(in *edgeproto.Platform, cb edgeproto.PlatformApi_ShowPlatformServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Platform) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
