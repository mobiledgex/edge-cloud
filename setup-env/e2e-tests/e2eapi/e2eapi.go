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
