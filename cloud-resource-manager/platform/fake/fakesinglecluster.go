package fake

import "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"

type PlatformSingleCluster struct {
	Platform
}

func (s *PlatformSingleCluster) GetFeatures() *platform.Features {
	features := s.Platform.GetFeatures()
	features.IsSingleKubernetesCluster = true
	features.SupportsAppInstDedicatedIP = true
	return features
}
