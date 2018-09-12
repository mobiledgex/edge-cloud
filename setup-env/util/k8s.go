package util

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
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
	if err != nil {
		return err
	}
	return nil
}

//convert the error to something easier to display on the test summary
func summarizeError(output string) string {
	if strings.Contains(output, "You do not currently have an active account selected") ||
		strings.Contains(output, "gcloud account not logged in") {
		return "need to run gcloud auth application-default login"
	}
	log.Printf("unknown error, no summary %s\n", output)
	return "unknown error - see logs"
}

func CreateK8sCluster() error {
	return runCommand("create", "mex", "cluster", "create", "-manifest", Deployment.Cluster.MexManifest)
}

func DeleteK8sCluster() error {
	return runCommand("delete", "mex", "cluster", "remove", "-manifest", Deployment.Cluster.MexManifest)
}

func runCommand(actionType string, command string, args ...string) error {
	log.Printf("runCommand %s args %v\n", command, args)
	cmd := exec.Command(command, args...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("Got error from k8s command %s args %v out: [%s] err: %v\n", command, args, output, err)
		if actionType == "create" && strings.Contains(string(output), "already exists") {
			log.Println("Ignoring already exists error")
		} else if actionType == "delete" && (strings.Contains(string(output), "not found") ||
			(strings.Contains(string(output), "no matches for kind"))) {
			log.Println("Ignoring not found error")
		} else {
			return errors.New(summarizeError(string(output)))
		}
	} else {
		log.Printf("K8s command OK: %s\n", output)
	}
	return nil
}

func checkPod(namespace string, podname string, count int, maxwait int) bool {
	log.Printf("check pod for name %s\n", podname)

	for i := 0; i <= maxwait; i++ {
		podlist, err := kubeClientset.CoreV1().Pods(namespace).List(v1.ListOptions{})

		readyCount := 0
		if err != nil {
			log.Printf("Error in list pods: %+v\n", err)
			return false
		}
		for _, p := range podlist.Items {
			if strings.HasPrefix(p.Name, podname) {
				log.Printf("Pod Name: %s status %v\n", p.Name, p.Status.Phase)
				if p.Status.Phase == core.PodRunning {
					readyCount += 1
				}
			}
		}
		if readyCount >= count {
			return true
		}
		log.Printf("waiting for pods: %d max %d\n", i, maxwait)
		time.Sleep(1 * time.Second)
	}
	return false
}

func copyFileToPods(podName string, src string, dest string) error {
	podlist, err := kubeClientset.CoreV1().Pods("default").List(v1.ListOptions{})
	if err != nil {
		return err
	}
	for _, p := range podlist.Items {
		if strings.HasPrefix(p.Name, podName) {
			err := runCommand("copy", "kubectl", "cp", src, p.Name+":"+dest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func DeployK8sServices(directory string) error {
	err := getKubeClient()
	if err != nil {
		return err
	}
	for _, k := range Deployment.K8sDeployment {
		err := runCommand("create", "kubectl", "create", "-f", directory+"/"+k.File)
		if err != nil {
			return err
		}
		for _, w := range k.WaitForPods {
			checkPod("default", w.PodName, w.PodCount, w.MaxWait)
			if err != nil {
				return err
			}
		}
		if len(k.CopyFiles) > 0 {
			for _, f := range k.CopyFiles {
				log.Printf("copying file: %s to %s\n", f.Src, f.Dest)
				err := copyFileToPods(f.PodName, directory+"/"+f.Src, f.Dest)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func DeleteK8sServices(directory string) error {
	//go thru in reverse order
	for i := len(Deployment.K8sDeployment) - 1; i >= 0; i-- {
		err := runCommand("delete", "kubectl", "delete", "-f", directory+"/"+Deployment.K8sDeployment[i].File)
		if err != nil {
			return err
		}
	}
	return nil
}

//get an ip and port for the service.  Wait up to maxwait for this to
//appear this takes a little while when the cluster is first deployed
func GetK8sServiceAddr(svcname string, maxwait int) (string, error) {
	err := getKubeClient()

	if err != nil {
		return "", err
	}
	for i := 0; i <= maxwait; i += 3 {
		//todo we may not always want default namespace
		svc, err := kubeClientset.CoreV1().Services("default").Get(svcname, v1.GetOptions{})
		if err != nil {
			return "", err
		}
		if len(svc.Status.LoadBalancer.Ingress) != 0 && len(svc.Spec.Ports) != 0 {
			//can be IP or DNS
			host := svc.Status.LoadBalancer.Ingress[0].IP
			port := fmt.Sprintf("%d", svc.Spec.Ports[0].Port)
			if host == "" {
				host = svc.Status.LoadBalancer.Ingress[0].Hostname
			}
			// sleep one more time after external ip shows up just to make sure it's ready before hitting it with traffic
			time.Sleep(1 * time.Second)
			for i := 0; i < 30; i++ {
				//wait for port to be reachable
				log.Printf("Check if port is open: %s\n", host+":"+port)
				conn, err := net.DialTimeout("tcp", host+":"+port, 1*time.Minute)
				if err == nil {
					conn.Close()
					log.Printf("Port is open: %s\n", host+":"+port)
					return fmt.Sprintf(host + ":" + port), nil
				}
				log.Printf("port not yet reachable: %v", err)
				time.Sleep(1 * time.Second)
			}
			return "", fmt.Errorf("Service created but not reachable")
		}
		log.Printf("waiting for service: %s iter %d max %d current status %+v \n", svcname, i, maxwait, svc.Status)
		time.Sleep(3 * time.Second)
	}

	return "", errors.New("unable to get service " + svcname)

}
