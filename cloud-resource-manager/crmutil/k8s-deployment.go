package crmutil

import (
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

//Kubeconfig contains kubernetes config file path
var Kubeconfig *string

func init() {
	if home := homedir.HomeDir(); home != "" {
		Kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		Kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

func getClientSet() (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", *Kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func constructDeploymentManifest(app *edgeproto.EdgeCloudApp) (*appsv1.Deployment, error) {
	if app.Name == "" {
		return nil, fmt.Errorf("Missing name")
	}

	if app.Image == "" {
		return nil, fmt.Errorf("Missing image specification")
	}

	// appname:version
	if !strings.Contains(app.Image, ":") {
		return nil, fmt.Errorf("Invalid image specification")
	}

	if app.Exposure == "" {
		return nil, fmt.Errorf("Missing exposure specification")
	}

	if app.Namespace == "" {
		return nil, fmt.Errorf("Missing namespace")
	}

	// portname,port
	// TODO expand the syntax, allow more ports
	exposures := strings.Split(app.Exposure, ",")
	if len(exposures) < 2 {
		return nil, fmt.Errorf("Invalid exposure specification")
	}
	portname := exposures[0]
	port, err := strconv.Atoi(exposures[1])
	if err != nil {
		return nil, fmt.Errorf("Cannot parse exposure specification")
	}

	if app.Replicas == 0 {
		app.Replicas = 1
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: app.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &app.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": app.Name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": app.Name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  app.Name,
							Image: app.Image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          portname,
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: int32(port),
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment, nil
}

func createKubernetesDeployment(app *edgeproto.EdgeCloudApp) error {
	deployment, err := constructDeploymentManifest(app)
	if err != nil {
		return err
	}

	clientset, err := getClientSet()
	if err != nil {
		return fmt.Errorf("cannot get clientset")
	}

	deploymentsClient := clientset.AppsV1().Deployments(app.Namespace)

	_, err = deploymentsClient.Create(deployment)
	return err
}

func deleteKubernetesDeployment(app *edgeproto.EdgeCloudApp) error {
	clientset, err := getClientSet()
	if err != nil {
		return fmt.Errorf("cannot get clientset")
	}

	deploymentsClient := clientset.AppsV1().Deployments(app.Namespace)

	err = deploymentsClient.Delete(app.Name, &metav1.DeleteOptions{})
	return err
}
