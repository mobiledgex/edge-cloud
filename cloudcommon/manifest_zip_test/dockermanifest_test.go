package ziptest

import (
	"archive/zip"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	yaml "github.com/mobiledgex/yaml/v2"
	"github.com/stretchr/testify/require"
)

func zipManifests() error {
	// Get a Buffer to Write To
	zipfile := "/tmp/temp.zip"
	outFile, err := os.Create(zipfile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)

	content, err := ioutil.ReadFile("manifest.yml")
	if err != nil {
		return fmt.Errorf("unable to open %s manifest file: %v", "manifest.yml", err)
	}
	var dm cloudcommon.DockerManifest
	err = yaml.Unmarshal([]byte(content), &dm)
	if err != nil {
		return fmt.Errorf("unmarshalling manifest.yml: %v", err)
	}
	zipFiles := []string{"manifest.yml"}
	zipFiles = append(zipFiles, dm.DockerComposeFiles...)
	for _, composeFile := range zipFiles {
		fileName := composeFile
		f, err := w.Create(fileName)
		if err != nil {
			return err
		}
		content, err = ioutil.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("unable to open %s file: %v", fileName, err)
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			return err
		}
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

var imagePaths = map[string]struct{}{
	"docker.mobiledgex.net/test/images/mysql:5.7":                 struct{}{},
	"docker-internal.mobiledgex.net/test/images/wordpress:latest": struct{}{},
	"docker-int.mobiledgex.net/test/images/mysql:5.7":             struct{}{},
	"docker-ext.mobiledgex.net/test/images/pgdb:1.1":              struct{}{},
}

func TestRemoteZipManifests(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// Create a zip file for testing
	err := zipManifests()
	require.Nil(t, err)

	// Zip file will be cleaned up by the following function
	zipContainers, err := cloudcommon.GetRemoteZipDockerManifests(ctx, nil, "", "", cloudcommon.NoDownload)
	require.Nil(t, err)
	for _, containers := range zipContainers {
		for _, container := range containers {
			_, ok := imagePaths[container.Image]
			require.True(t, ok, "valid image path")
		}
	}
}
