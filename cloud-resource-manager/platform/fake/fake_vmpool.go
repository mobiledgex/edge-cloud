package fake

import "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"

type PlatformVMPool struct {
	Platform
}

func (s *PlatformVMPool) GetFeatures() *platform.Features {
	features := s.Platform.GetFeatures()
	features.IsVMPool = true
	return features
}
