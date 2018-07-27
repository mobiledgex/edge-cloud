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

//RunApp runs the cloud app
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
	case "openstack-simple":
		if len(app.Apps) < 1 {
			return fmt.Errorf("no apps specified")
		}
		app0 := app.Apps[0]
		client, err := GetOpenstackClient(app0.Region)
		if err != nil {
			return err
		}

		args := OpenstackServerArgs{
			Name:    app0.Name,
			Region:  app0.Region,
			Image:   app0.Image,
			Flavor:  app0.Flavor,
			Network: app0.Network,
		}
		if err := CreateOpenstackServer(client, &args); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Invalid app Kind")
	}

	return nil
}

//KillApp stops and deletes cloud app
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
	case "openstack-simple":
		if len(app.Apps) < 1 {
			return fmt.Errorf("no apps specified")
		}
		app0 := app.Apps[0]
		client, err := GetOpenstackClient(app0.Region)
		if err != nil {
			return err
		}

		if err := DeleteOpenstackServer(client, app0.Name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Invalid app Kind")
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
	defer func() {
		err = out.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func getTempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

//TODO UpdateApp
