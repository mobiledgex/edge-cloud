package crmutil

import (
	"os"

	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/azure"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/gcloud"
	log "gitlab.com/bobbae/logrus"
)

//ClusterCreate creates a cluster
func ClusterCreate(mf *Manifest) {
	log.Debugf("creating cluster, %v", mf)

	switch mf.Kind {
	case "mex-openstack-kubernetes":
		if err := CreateCluster(mf.Spec.RootLB, mf.Spec.Flavor, mf.Metadata.Name,
			mf.Spec.Networks[0].Kind+","+mf.Spec.Networks[0].Name+","+mf.Spec.Networks[0].CIDR,
			mf.Metadata.Tags, mf.Metadata.Tenant); err != nil {
			log.Fatalf("can't create cluster, %v", err)
		}
	case "gcloud-gke":
		if err := gcloud.SetProject(mf.Metadata.Project); err != nil {
			log.Fatal(err)
		}
		if err := gcloud.SetZone(mf.Metadata.Zone); err != nil {
			log.Fatal(err)
		}
		if err := gcloud.CreateGKECluster(mf.Metadata.Name); err != nil {
			log.Fatal(err)
		}
		if err := gcloud.GetGKECredentials(mf.Metadata.Name); err != nil {
			log.Fatal(err)
		}
	case "azure-aks":
		if err := azure.CreateResourceGroup(mf.Metadata.ResourceGroup, mf.Metadata.Location); err != nil {
			log.Fatal(err)
		}
		if err := azure.CreateAKSCluster(mf.Metadata.ResourceGroup, mf.Metadata.Location); err != nil {
			log.Fatal(err)
		}
		if err := azure.GetAKSCredentials(mf.Metadata.ResourceGroup, mf.Metadata.Location); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("created cluster", mf)
}

//ClusterRemove removes a cluster
func ClusterRemove(mf *Manifest) {
	log.Debugf("removing cluster, %v", mf)

	switch mf.Kind {
	case "mex-openstack-kubernetes":
		if err := DeleteClusterByName(mf.Spec.RootLB, mf.Metadata.Name); err != nil {
			log.Fatalf("can't remove cluster, %v", err)
		}
	case "gcloud-gke":
		if err := gcloud.DeleteGKECluster(mf.Metadata.Name); err != nil {
			log.Fatal(err)
		}
	case "azure-aks":
		if err := azure.DeleteAKSCluster(mf.Metadata.ResourceGroup); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("removed cluster", mf)
}

//SetEnvVars sets up environment vars and checks for credentials required for running
func SetEnvVars(mf *Manifest) {
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

	if err := os.Setenv("MEX_ROOT_LB", mf.Metadata.Name); err != nil {
		log.Fatal(err)
	}
	if err := os.Setenv("MEX_AGENT_IMAGE", mf.Spec.Agent.Image); err != nil {
		log.Fatal(err)
	}
	if err := os.Setenv("MEX_ZONE", mf.Metadata.DNSZone); err != nil {
		log.Fatal(err)
	}
	if err := os.Setenv("MEX_EXT_NETWORK", mf.Spec.ExternalNetwork); err != nil {
		log.Fatal(err)
	}
	if err := os.Setenv("MEX_NETWORK", mf.Spec.InternalNetwork); err != nil {
		log.Fatal(err)
	}
	if err := os.Setenv("MEX_EXT_ROUTER", mf.Spec.ExternalRouter); err != nil {
		log.Fatal(err)
	}
	if err := os.Setenv("MEX_DOCKER_REGISTRY", mf.Spec.DockerRegistry); err != nil {
		log.Fatal(err)
	}

	log.Debugln("MEX_ROOT_LB", mf.Metadata.Name)
	log.Debugln("MEX_AGENT_IMAGE", mf.Spec.Agent.Image)
	log.Debugln("MEX_ZONE", mf.Metadata.DNSZone)
	log.Debugln("MEX_EXT_NETWORK", mf.Spec.ExternalNetwork)
	log.Debugln("MEX_NETWORK", mf.Spec.InternalNetwork)
	log.Debugln("MEX_EXT_ROUTER", mf.Spec.ExternalRouter)
	log.Debugln("MEX_DOCKER_REGISTRY", mf.Spec.DockerRegistry)
}

//PlatformInit initializes platform
func PlatformInit(mf *Manifest) {
	log.Debugf("init platform, %v", mf)

	switch mf.Kind {
	case "mex-openstack-kubernetes":
		SetEnvVars(mf)
		if err := RunMEXAgent(mf.Metadata.Name, false); err != nil {
			log.Fatal(err)
		}
	case "gcloud-gke":
	case "azure-aks":
	}
}

//PlatformClean cleans up the platform
func PlatformClean(mf *Manifest) {
	log.Debugf("clean platform, %v", mf)

	switch mf.Kind {

	case "mex-openstack-kubernetes":
		SetEnvVars(mf)
		if err := RemoveMEXAgent(mf.Metadata.Name); err != nil {
			log.Fatal(err)
		}
	case "gcloud-gke":
	case "azure-aks":
	}
}
