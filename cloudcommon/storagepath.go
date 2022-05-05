// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloudcommon

import (
	"fmt"
	"strings"

	"github.com/edgexr/edge-cloud/util"
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
