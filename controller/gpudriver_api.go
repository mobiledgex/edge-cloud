package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/gcs"
	"github.com/mobiledgex/edge-cloud/log"
)

type GPUDriverApi struct {
	sync  *Sync
	store edgeproto.GPUDriverStore
	cache edgeproto.GPUDriverCache
}

var gpuDriverApi = GPUDriverApi{}

func InitGPUDriverApi(sync *Sync) {
	gpuDriverApi.sync = sync
	gpuDriverApi.store = edgeproto.NewGPUDriverStore(sync.store)
	edgeproto.InitGPUDriverCache(&gpuDriverApi.cache)
	sync.RegisterCache(&gpuDriverApi.cache)
}

func getGCSStorageClient(ctx context.Context) (*gcs.GCSClient, error) {
	bucketName := cloudcommon.GetGPUDriverBucketName(nodeMgr.DeploymentTag)
	credsObj, err := gcs.GetGCSCreds(ctx, vaultConfig)
	if err != nil {
		return nil, err
	}
	storageClient, err := gcs.NewClient(ctx, credsObj, bucketName, gcs.LongTimeout)
	if err != nil {
		return nil, fmt.Errorf("Unable to setup GCS client: %v", err)
	}
	return storageClient, nil
}

func setupGPUDriver(ctx context.Context, storageClient *gcs.GCSClient, driverKey *edgeproto.GPUDriverKey, build *edgeproto.GPUDriverBuild) error {
	if build.DriverPath == "" {
		return fmt.Errorf("Missing driverpath: %s", build.Name)
	}
	if build.OperatingSystem == edgeproto.OSType_LINUX && build.KernelVersion == "" {
		return fmt.Errorf("Kernel version is required for Linux build %s", build.Name)
	}
	driverFileName, err := cloudcommon.GetFileNameWithExt(build.DriverPath)
	if err != nil {
		return err
	}
	ext := filepath.Ext(driverFileName)
	// Download the package
	var fileContents string
	authApi := &cloudcommon.VaultRegistryAuthApi{
		VaultConfig: vaultConfig,
	}
	err = cloudcommon.DownloadFile(ctx, authApi, build.DriverPath, "", &fileContents)
	if err != nil {
		return fmt.Errorf("Failed to download GPU driver build %s, %v", build.DriverPath, err)
	}
	// TODO: If linux, then validate the pkg
	//                  * Pkg must be deb pkg
	//                  * Pkg control file must depend on specified kernel version

	// Upload the package to GCS
	fileName := cloudcommon.GetGPUDriverBuildStoragePath(driverKey, build.Name, ext)
	err = storageClient.UploadObject(ctx, fileName, bytes.NewBufferString(fileContents))
	if err != nil {
		return fmt.Errorf("Failed to upload GPU driver build to %s, %v", fileName, err)
	}
	build.DriverPath = cloudcommon.GetGPUDriverURL(driverKey, nodeMgr.DeploymentTag, build.Name, ext)
	return nil
}

func setupGPUDriverLicenseConfig(ctx context.Context, storageClient *gcs.GCSClient, driverKey *edgeproto.GPUDriverKey, licenseConfig *string) error {
	fileName := cloudcommon.GetGPUDriverLicenseStoragePath(driverKey)
	err := storageClient.UploadObject(ctx, fileName, bytes.NewBufferString(*licenseConfig))
	if err != nil {
		return fmt.Errorf("Failed to upload GPU driver license to %s, %v", fileName, err)
	}
	*licenseConfig = cloudcommon.GetGPUDriverLicenseURL(driverKey, nodeMgr.DeploymentTag)
	return nil
}

func deleteGPUDriverLicenseConfig(ctx context.Context, storageClient *gcs.GCSClient, driverKey *edgeproto.GPUDriverKey) error {
	fileName := cloudcommon.GetGPUDriverLicenseStoragePath(driverKey)
	err := storageClient.DeleteObject(ctx, fileName)
	if err != nil {
		return fmt.Errorf("Failed to delete GPU driver license to %s, %v", fileName, err)
	}
	return nil
}

func (s *GPUDriverApi) CreateGPUDriver(ctx context.Context, in *edgeproto.GPUDriver) (retres *edgeproto.Result, reterr error) {
	if err := in.Validate(nil); err != nil {
		return &edgeproto.Result{}, err
	}

	if in.Key.Organization == "" {
		// By default GPU drivers are owned by MobiledgeX org
		in.Key.Organization = cloudcommon.OrganizationMobiledgeX
	}

	// Step-1: First commit to etcd
	// Step-2: Validate and upload the builds/license-config to GCS
	// Step-3: And then update build details to reflect GCS URL and update it to etcd

	// Step-1: First commit to etcd
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}

	defer func() {
		if reterr != nil {
			// undo changes
			_, err = s.DeleteGPUDriver(ctx, in)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelApi, "failed to undo gpu driver create", "key", in.Key, "err", err)
			}
		}
	}()

	storageClient, err := getGCSStorageClient(ctx)
	if err != nil {
		return nil, err
	}
	defer storageClient.Close()

	// Step-2: Validate and upload the builds/license-config to GCS
	for ii, build := range in.Builds {
		err := setupGPUDriver(ctx, storageClient, &in.Key, &build)
		if err != nil {
			return nil, err
		}
		in.Builds[ii].DriverPath = build.DriverPath
	}

	// If license config is present, upload it to GCS
	if in.LicenseConfig != "" {
		err := setupGPUDriverLicenseConfig(ctx, storageClient, &in.Key, &in.LicenseConfig)
		if err != nil {
			return nil, err
		}
	}

	// Step-3: And then update build details to reflect GCS URL and update it to etcd
	_, err = s.store.Put(ctx, in, s.sync.syncWait)
	return &edgeproto.Result{}, err
}

func (s *GPUDriverApi) UpdateGPUDriver(ctx context.Context, in *edgeproto.GPUDriver) (*edgeproto.Result, error) {
	err := in.ValidateUpdateFields()
	if err != nil {
		return &edgeproto.Result{}, err
	}
	fmap := edgeproto.MakeFieldMap(in.Fields)
	if err := in.Validate(fmap); err != nil {
		return &edgeproto.Result{}, err
	}

	// If license config is present, upload it to GCS
	if _, found := fmap[edgeproto.GPUDriverFieldLicenseConfig]; found {
		if in.LicenseConfig == "" {
			return nil, fmt.Errorf("LicenseConfig cannot be empty")
		}
		storageClient, err := getGCSStorageClient(ctx)
		if err != nil {
			return nil, err
		}
		defer storageClient.Close()
		err = setupGPUDriverLicenseConfig(ctx, storageClient, &in.Key, &in.LicenseConfig)
		if err != nil {
			return nil, err
		}
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.GPUDriver{}
		changed := 0
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed = cur.CopyInFields(in)
		if err := cur.Validate(nil); err != nil {
			return err
		}
		if changed == 0 {
			return nil
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *GPUDriverApi) DeleteGPUDriver(ctx context.Context, in *edgeproto.GPUDriver) (*edgeproto.Result, error) {
	if err := in.Key.ValidateKey(); err != nil {
		return &edgeproto.Result{}, err
	}
	storageClient, err := getGCSStorageClient(ctx)
	if err != nil {
		return nil, err
	}
	defer storageClient.Close()
	buildFiles := []string{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.NotFoundError()
		}
		buildFiles = []string{}
		for _, build := range in.Builds {
			fileName := cloudcommon.GetGPUDriverBuildPathFromURL(build.DriverPath, nodeMgr.DeploymentTag)
			buildFiles = append(buildFiles, fileName)
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}
	// Delete license config from GCS
	err = deleteGPUDriverLicenseConfig(ctx, storageClient, &in.Key)
	if err != nil {
		return nil, err
	}
	// Delete builds from GCS
	for _, filename := range buildFiles {
		err = storageClient.DeleteObject(ctx, filename)
		if err != nil {
			return nil, fmt.Errorf("Failed to delete GPU driver build %s, %v: %v", filename, in.Key, err)
		}
	}
	return &edgeproto.Result{}, nil
}

func (s *GPUDriverApi) ShowGPUDriver(in *edgeproto.GPUDriver, cb edgeproto.GPUDriverApi_ShowGPUDriverServer) error {
	return s.cache.Show(in, func(obj *edgeproto.GPUDriver) error {
		err := cb.Send(obj)
		return err
	})
}

func (s *GPUDriverApi) AddGPUDriverBuild(ctx context.Context, in *edgeproto.GPUDriverBuildMember) (retres *edgeproto.Result, reterr error) {
	if err := in.Validate(); err != nil {
		return &edgeproto.Result{}, err
	}
	// Step-1: First commit to etcd
	// Step-2: Validate and upload the build to GCS
	// Step-3: And then update build details to reflect GCS URL and update it to etcd

	// Step-1: First commit to etcd
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.GPUDriver{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		for ii, _ := range cur.Builds {
			if cur.Builds[ii].Name == in.Build.Name {
				return fmt.Errorf("GPU driver build with same name already exists")
			}
		}
		cur.Builds = append(cur.Builds, in.Build)
		s.store.STMPut(stm, &cur)
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}

	defer func() {
		if reterr != nil {
			// undo changes
			_, err = s.RemoveGPUDriverBuild(ctx, in)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelApi, "failed to undo gpu driver build", "key", in.Key, "err", err)
			}
		}
	}()

	// Step-2: Validate and upload the build to GCS
	storageClient, err := getGCSStorageClient(ctx)
	if err != nil {
		return nil, err
	}
	defer storageClient.Close()

	err = setupGPUDriver(ctx, storageClient, &in.Key, &in.Build)
	if err != nil {
		return nil, err
	}

	// Step-3: And then update build details to reflect GCS URL and update it to etcd
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.GPUDriver{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		found := false
		for ii, _ := range cur.Builds {
			if cur.Builds[ii].Name == in.Build.Name {
				cur.Builds[ii] = in.Build
				found = true
				break
			}
		}
		if !found {
			cur.Builds = append(cur.Builds, in.Build)
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *GPUDriverApi) RemoveGPUDriverBuild(ctx context.Context, in *edgeproto.GPUDriverBuildMember) (*edgeproto.Result, error) {
	if err := in.Key.ValidateKey(); err != nil {
		return &edgeproto.Result{}, err
	}
	if err := in.Build.ValidateName(); err != nil {
		return &edgeproto.Result{}, err
	}
	driverURL := ""
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.GPUDriver{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed := false
		for ii, build := range cur.Builds {
			if build.Name == in.Build.Name {
				cur.Builds = append(cur.Builds[:ii], cur.Builds[ii+1:]...)
				driverURL = build.DriverPath
				changed = true
				break
			}
		}
		if !changed {
			return nil
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}
	if driverURL == "" {
		return &edgeproto.Result{}, nil
	}
	// Delete build from GCS
	fileName := cloudcommon.GetGPUDriverBuildPathFromURL(driverURL, nodeMgr.DeploymentTag)
	storageClient, err := getGCSStorageClient(ctx)
	if err != nil {
		return nil, err
	}
	defer storageClient.Close()
	err = storageClient.DeleteObject(ctx, fileName)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to delete GCS object", "filename", fileName, "err", err)
		return nil, err
	}
	return &edgeproto.Result{}, nil
}

func (s *GPUDriverApi) GetCloudletGPUDrivers(gpuType edgeproto.GPUType, driverName, cloudletOrg string) ([]edgeproto.GPUDriver, error) {
	if driverName == "" {
		return nil, fmt.Errorf("Missing GPU driver name")
	}
	gpuDriver := edgeproto.GPUDriver{
		Key: edgeproto.GPUDriverKey{
			Name: driverName,
			Type: gpuType,
		},
	}
	gpuDrivers := []edgeproto.GPUDriver{}
	err := s.cache.Show(&gpuDriver, func(obj *edgeproto.GPUDriver) error {
		if obj.Key.Organization == cloudcommon.OrganizationMobiledgeX ||
			obj.Key.Organization == cloudletOrg {
			gpuDrivers = append(gpuDrivers, *obj)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to get list of gpu drivers: %v", err)
	}
	return gpuDrivers, nil
}
