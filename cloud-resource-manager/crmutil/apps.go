package crmutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func RunApp(app *edgeproto.EdgeCloudApplication) error {
	switch app.Kind {
	case "k8s-manifest":
		df := getManifest(app.Manifest)
		if df != "" {
			if err := runKubeCtlCreateDeployment(df); err != nil {
				return err
			}
			return nil
		}
	case "k8s-simple":
		if len(app.Apps) < 1 {
			return fmt.Errorf("no apps specified")
		}
		if err := createKubernetesDeployment(app.Apps[0]); err != nil {
			return err
		}
	default:
	}

	return nil
}

func KillApp(app *edgeproto.EdgeCloudApplication) error {
	switch app.Kind {
	case "k8s-manifest":
		df := getManifest(app.Manifest)
		if df != "" {
			if err := runKubeCtlDeleteDeployment(df); err != nil {
				return err
			}
			return nil
		}
	case "k8s-simple":
		if len(app.Apps) < 1 {
			return fmt.Errorf("no apps specified")
		}
		if err := deleteKubernetesDeployment(app.Apps[0]); err != nil {
			return err
		}
	default:
	}
	return nil
}

func getManifest(manifest string) string {
	if manifest == "" {
		return ""
	}
	if strings.HasPrefix(manifest, "http") {
		localFile := getTempFileName("crm-", ".yaml")
		if err := getHTTPFile(manifest, localFile); err != nil {
			log.Printf("Can't get %s over HTTP, %v", localFile, err)
			return ""
		}
		return localFile
	}
	return ""
}

func getHTTPFile(uri string, localFile string) error {
	out, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func getTempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

//TODO UpdateApp
