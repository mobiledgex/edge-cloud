package k8s

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeClientset *kubernetes.Clientset
var kubeConfig *restclient.Config

func getKubeClient() error {

	home := os.Getenv("HOME")
	var err error

	kubeConfig, err = clientcmd.BuildConfigFromFlags("", home+"/.kube/config")

	if err != nil {
		return err
	}
	kubeClientset, err = kubernetes.NewForConfig(kubeConfig)
	log.Printf("Clientset %+v Err %v\n", kubeClientset, err)
	if err != nil {
		return err
	}
	return nil
}

//convert the error to something easier to display on the test summary
func summarizeError(output string) string {
	if strings.Contains(output, "You do not currently have an active account selected") {
		return "gcloud account not logged in"
	}
	return "unknown gcloud error - see logs"
}

func createCluster() error {
	return runGCloudCommand("create", "gcloud", "container", "clusters", "create", util.Deployment.GCloud.Cluster, "--zone", util.Deployment.GCloud.Zone, "--machine-type", util.Deployment.GCloud.MachineType)
}

func deleteCluster() error {
	return runGCloudCommand("delete", "gcloud", "container", "clusters", "delete", util.Deployment.GCloud.Cluster, "--quiet")
}

func runGCloudCommand(actionType string, command string, args ...string) error {
	log.Printf("runGCloudCommand for command %s args %v\n", command, args)
	cmd := exec.Command(command, args...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("Got error from k8s command %s args %v out: [%s] err: %v\n", command, args, output, err)
		if actionType == "create" && strings.Contains(string(output), "already exists") {
			log.Println("Ignoring already exists error")
		} else if actionType == "delete" && strings.Contains(string(output), "not found") {
			log.Println("Ignoring not found error")
		} else {
			return errors.New(summarizeError(string(output)))
		}
	} else {
		log.Printf("K8s command OK: %s\n", output)
	}
	return nil
}

func checkService(namespace string, svcname string) bool {
	svclist, err := kubeClientset.CoreV1().Services(namespace).List(v1.ListOptions{})
	if err != nil {
		log.Printf("Error in list services: %+v\n", err)
		return false
	}
	for _, s := range svclist.Items {
		log.Printf("SVC is: %+v\n", s)
	}
	return true
}

func DeployGoogleK8s(directory string) error {
	err := getKubeClient()
	if err != nil {
		return err
	}
	err = createCluster()
	if err != nil {
		return err
	}
	for _, k := range util.Deployment.K8sDeployment {
		//kubeClientset.AppsV1().
		err := runGCloudCommand("create", "kubectl", "create", "-f", directory+"/"+k.File)
		checkService("default", "etcd-operator")
		if err != nil {
			return err
		}
	}
	return nil
}

func CleanupGoogleK8s(directory string) error {
	//go thru in reverse order
	for i := len(util.Deployment.K8sDeployment) - 1; i >= 0; i-- {
		err := runGCloudCommand("delete", "kubectl", "delete", "-f", directory+"/"+util.Deployment.K8sDeployment[i].File)
		if err != nil {
			return err
		}
	}
	return deleteCluster()
}
