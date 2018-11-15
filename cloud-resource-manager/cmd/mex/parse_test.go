package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/stretchr/testify/require"
)

// test that parser works on checked-in files
// this catches changes to code that would break parsing existing yaml files
func TestParse(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	cwd := filepath.Dir(filename)

	files, err := ioutil.ReadDir(cwd)
	require.Nil(t, err, "read dir %s", cwd)
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".yaml") {
			continue
		}
		fmt.Printf("test parse %s\n", f.Name())
		data, err := ioutil.ReadFile(f.Name())
		require.Nil(t, err, "read file %s", f.Name())
		mf := &crmutil.Manifest{}
		err = yaml.Unmarshal(data, mf)
		require.Nil(t, err, "unmarshal manifest")
	}
}
