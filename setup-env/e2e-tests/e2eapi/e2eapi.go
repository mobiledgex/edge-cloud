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

package e2eapi

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// e2e-tests passes information to a test program to run a test.
// The "config" data it passes is defined below.
// The "testspec" data it passes is just a json string which
// e2e-tests does not understand, such that test programs can
// implement their own functionality and test fields without
// e2e-tests' knowledge.

type TestConfig struct {
	SetupFile string
	Vars      map[string]string
}

func (tc *TestConfig) Validate() []error {
	errs := make([]error, 0)

	if tc.SetupFile == "" {
		errs = append(errs, fmt.Errorf("missing SetupFile"))
	} else {
		if _, err := os.Stat(tc.SetupFile); err != nil {
			errs = append(errs, fmt.Errorf("SetupFile "+tc.SetupFile+" does not exist"))
		}
	}
	// outputdir is special, always required
	if _, found := tc.Vars["outputdir"]; !found {
		errs = append(errs, fmt.Errorf("outputdir not found in Vars"))
	}

	return errs
}

func ReadVarsFile(varsFile string, vars map[string]string) error {
	if varsFile == "" {
		return nil
	}
	dat, err := ioutil.ReadFile(varsFile)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(dat, &vars)
}
