package cloudcommon

import (
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/util"
)

type StoragePath struct {
	path []string
}

func (s *StoragePath) AppendPaths(paths ...string) error {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if err := util.ValidStoragePath(p); err != nil {
			return err
		}
		s.path = append(s.path, p)
	}
	err := s.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (s *StoragePath) Validate() error {
	if len(s.path) > 1 {
		if s.path[0] == ".well-known" && s.path[1] == "acme-challenge" {
			return fmt.Errorf("Path cannot start with '.well-known/acme-challenge/'")
		}
	}
	return nil
}

func (s *StoragePath) String() string {
	return strings.Join(s.path, "/")
}
