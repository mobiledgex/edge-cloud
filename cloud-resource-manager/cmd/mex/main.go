package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	log "gitlab.com/bobbae/logrus"

	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/azure"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/gcloud"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
)

type resource struct {
	name    string
	handler func(args []string)
}

var resources = []resource{
	{"cluster", clusterHandler},
	{"platform", platformHandler},
	//{"network", networkHandler},
	//{"storage", storageHandler},
	//{"configmap", configmapHandler},
	//{"deployment", deploymentHandler},
	//{"service", serviceHandler},
	//{"namespace", namespaceHandler},
}

func main() {
	debug := flag.Bool("debug", false, "debugging")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	args := flag.Args()

	log.Debugln("args", args)

	if len(args) < 2 {
		log.Infoln("example1: mex cluster create -manifest mycluster.yaml")
		log.Infoln("example2: mex cluster remove -manifest mycluster.yaml")
		log.Fatalf("insufficient args")
	}
	found := 0
	for _, r := range resources {
		if args[0] == r.name {
			r.handler(args[1:])
			found++
		}
	}
	if found == 0 {
		rlist := []string{}
		for _, r := range resources {
			rlist = append(rlist, r.name)
		}
		log.Fatalf("valid resources are: %v", rlist)
	}
}

var clusterCommands = []string{"create", "remove"}
var platformCommands = []string{"init", "clean"}

func validateCommand(rsrc string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing command")
	}

	commands := []string{}

	switch rsrc {
	case "cluster":
		commands = clusterCommands
	case "platform":
		commands = platformCommands
	default:
		log.Fatalln("resource not handled yet")
	}

	cmd := args[0]
	for _, c := range commands {
		if c == cmd {
			return nil
		}
	}
	return fmt.Errorf("valid commands are %v", commands)
}

func clusterHandler(args []string) {
	subflags := flag.NewFlagSet("cluster", flag.ExitOnError)

	manifest := subflags.String("manifest", "", "manifest for cluster")

	if err := validateCommand("cluster", args); err != nil {
		log.Fatal(err)
	}

	cmd := args[0]
	args = args[1:]
	subflags.Parse(args)
	log.Debugln("cluster", cmd, args)

	cl := &crmutil.ClusterManifest{}
	dat, err := ioutil.ReadFile(*manifest)
	if err != nil {
		log.Fatalf("can't read %s, %v", *manifest, err)
	}
	err = yaml.Unmarshal(dat, cl)
	if err != nil {
		log.Fatalf("can't unmarshal, %v", err)
	}
	switch cmd {
	case "create":
		clusterCreate(cl)
	case "remove":
		clusterRemove(cl)
	}
}

func clusterCreate(cl *crmutil.ClusterManifest) {
	log.Debugf("creating cluster, %v", cl)

	switch cl.Kind {
	case "mex-openstack-kubernetes":
		if err := crmutil.CreateCluster(cl.Spec.RootLB, cl.Spec.Flavor, cl.Metadata.Name,
			cl.Spec.Networks[0].Kind+","+cl.Spec.Networks[0].Name+","+cl.Spec.Networks[0].CIDR,
			cl.Metadata.Tags, cl.Metadata.Tenant); err != nil {
			log.Fatalf("can't create cluster, %v", err)
		}
	case "gcloud-gke":
		if err := gcloud.SetProject(cl.Metadata.Project); err != nil {
			log.Fatal(err)
		}
		if err := gcloud.SetZone(cl.Metadata.Zone); err != nil {
			log.Fatal(err)
		}
		if err := gcloud.CreateGKECluster(cl.Metadata.Name); err != nil {
			log.Fatal(err)
		}
		if err := gcloud.GetGKECredentials(cl.Metadata.Name); err != nil {
			log.Fatal(err)
		}
	case "azure-aks":
		if err := azure.CreateResourceGroup(cl.Metadata.ResourceGroup, cl.Metadata.Location); err != nil {
			log.Fatal(err)
		}
		if err := azure.CreateAKSCluster(cl.Metadata.ResourceGroup, cl.Metadata.Location); err != nil {
			log.Fatal(err)
		}
		if err := azure.GetAKSCredentials(cl.Metadata.ResourceGroup, cl.Metadata.Location); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("created cluster", cl)
}

func clusterRemove(cl *crmutil.ClusterManifest) {
	log.Debugf("removing cluster, %v", cl)

	switch cl.Kind {
	case "mex-openstack-kubernetes":
		if err := crmutil.DeleteClusterByName(cl.Spec.RootLB, cl.Metadata.Name); err != nil {
			log.Fatalf("can't remove cluster, %v", err)
		}
	case "gcloud-gke":
		if err := gcloud.DeleteGKECluster(cl.Metadata.Name); err != nil {
			log.Fatal(err)
		}
	case "azure-aks":
		if err := azure.DeleteAKSCluster(cl.Metadata.ResourceGroup); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("removed cluster", cl)
}

func platformHandler(args []string) {
	subflags := flag.NewFlagSet("platform", flag.ExitOnError)

	manifest := subflags.String("manifest", "", "manifest for platform")

	if err := validateCommand("platform", args); err != nil {
		log.Fatal(err)
	}

	cmd := args[0]
	args = args[1:]
	subflags.Parse(args)
	log.Debugln("platform", cmd, args)

	pl := &crmutil.PlatformManifest{}
	dat, err := ioutil.ReadFile(*manifest)
	if err != nil {
		log.Fatalf("can't read %s, %v", *manifest, err)
	}
	err = yaml.Unmarshal(dat, pl)
	if err != nil {
		log.Fatalf("can't unmarshal, %v", err)
	}
	switch cmd {
	case "init":
		platformInit(pl)
	case "clean":
		platformClean(pl)
	}
}

func setEnvVars(pl *crmutil.PlatformManifest) {
	// secrets to be passed via Env var still : MEX_CF_KEY, MEX_CF_USER, MEX_DOCKER_REG_PASS
	// TODO: use `secrets` or `vault`

	eCFKey := os.Getenv("MEX_CF_KEY")
	if eCFKey == "" {
		log.Fatalln("no MEX_CF_KEY")
	}
	eCFUser := os.Getenv("MEX_CF_USER")
	if eCFUser == "" {
		log.Fatalln("no MEX_CF_USER")
	}
	eMEXDockerRegPass := os.Getenv("MEX_DOCKER_REG_PASS")
	if eMEXDockerRegPass == "" {
		log.Fatalln("no MEX_DOCKER_REG_PASS")
	}

	os.Setenv("MEX_ROOT_LB", pl.Metadata.Name)
	os.Setenv("MEX_AGENT_IMAGE", pl.Spec.Agent.Image)
	os.Setenv("MEX_ZONE", pl.Metadata.DNSZone)
	os.Setenv("MEX_EXT_NETWORK", pl.Spec.ExternalNetwork)
	os.Setenv("MEX_NETWORK", pl.Spec.InternalNetwork)
	os.Setenv("MEX_EXT_ROUTER", pl.Spec.ExternalRouter)
	os.Setenv("MEX_DOCKER_REGISTRY", pl.Spec.DockerRegistry)
}

func platformInit(pl *crmutil.PlatformManifest) {
	log.Debugf("init platform, %v", pl)

	switch pl.Kind {
	case "mex-openstack-kubernetes":
		setEnvVars(pl)
		if err := crmutil.RunMEXAgent(pl.Metadata.Name, false); err != nil {
			log.Fatal(err)
		}
	case "gcloud-gke":
	case "azure-aks":
	}
}

func platformClean(pl *crmutil.PlatformManifest) {
	log.Debugf("clean platform, %v", pl)

	switch pl.Kind {

	case "mex-openstack-kubernetes":
		setEnvVars(pl)
		if err := crmutil.RemoveMEXAgent(pl.Metadata.Name); err != nil {
			log.Fatal(err)
		}
	case "gcloud-gke":
	case "azure-aks":
	}
}
