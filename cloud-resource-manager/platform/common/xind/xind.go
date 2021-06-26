package xind

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

// Common code for DIND and KIND
type Xind struct {
	Caches         *platform.Caches
	remotePassword string
	clusterManager ClusterManager
}

func (s *Xind) Init(ctx context.Context, platformConfig *platform.PlatformConfig, caches *platform.Caches, clusterManager ClusterManager, updateCallback edgeproto.CacheUpdateCallback) error {
	s.Caches = caches
	s.clusterManager = clusterManager
	// for backwards compatibility, removes l7 proxies, can delete this later
	client, err := s.GetNodePlatformClient(ctx, nil)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Setting up Nginx L7 Proxy")
	err = proxy.InitL7Proxy(ctx, client, proxy.WithDockerPublishPorts())
	if err != nil {
		return err
	}
	return nil
}

func (s *Xind) GetFeatures() *platform.Features {
	return &platform.Features{
		CloudletServicesLocal: true,
	}
}

func (s *Xind) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	client, err := s.GetClient(ctx)
	if err != nil {
		return err
	}
	err = GetLimits(ctx, client, info)
	if err != nil {
		return err
	}
	// Use flavors from controller as platform flavor
	var flavors []*edgeproto.FlavorInfo
	if s.Caches == nil {
		return fmt.Errorf("Flavor cache is nil")
	}
	flavorkeys := make(map[edgeproto.FlavorKey]struct{})
	s.Caches.FlavorCache.GetAllKeys(ctx, func(k *edgeproto.FlavorKey, modRev int64) {
		flavorkeys[*k] = struct{}{}
	})
	for k := range flavorkeys {
		var flav edgeproto.Flavor
		if s.Caches.FlavorCache.Get(&k, &flav) {
			var flavInfo edgeproto.FlavorInfo
			flavInfo.Name = flav.Key.Name
			flavInfo.Ram = flav.Ram
			flavInfo.Vcpus = flav.Vcpus
			flavInfo.Disk = flav.Disk
			flavors = append(flavors, &flavInfo)
		} else {
			return fmt.Errorf("fail to fetch flavor %s", k)
		}
	}
	info.Flavors = flavors
	return nil
}

func (s *Xind) GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst, clientType string) (ssh.Client, error) {
	return s.GetClient(ctx)
}

func (s *Xind) GetNodePlatformClient(ctx context.Context, node *edgeproto.CloudletMgmtNode, ops ...pc.SSHClientOp) (ssh.Client, error) {
	return s.GetClient(ctx)
}

func (s *Xind) GetClient(ctx context.Context) (ssh.Client, error) {
	// TODO: add support for remote infra
	return &pc.LocalClient{
		WorkingDir: "/tmp",
	}, nil
}

func (s *Xind) ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst, vmAppInsts []edgeproto.AppInst) ([]edgeproto.CloudletMgmtNode, error) {
	return []edgeproto.CloudletMgmtNode{}, nil
}

func (s *Xind) GetCloudletProps(ctx context.Context) (*edgeproto.CloudletProps, error) {
	return &edgeproto.CloudletProps{}, nil
}

func (s *Xind) GetAccessData(ctx context.Context, cloudlet *edgeproto.Cloudlet, region string, vaultConfig *vault.Config, dataType string, arg []byte) (map[string]string, error) {
	return nil, nil
}

func (s *Xind) GetRootLBClients(ctx context.Context) (map[string]ssh.Client, error) {
	return nil, nil
}

func (s *Xind) GetVersionProperties() map[string]string {
	return map[string]string{}
}

func (s *Xind) GetRootLBFlavor(ctx context.Context) (*edgeproto.Flavor, error) {
	return &edgeproto.Flavor{
		Vcpus: uint64(0),
		Ram:   uint64(0),
		Disk:  uint64(0),
	}, nil
}
