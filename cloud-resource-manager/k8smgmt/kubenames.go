package k8smgmt

import (
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	v1 "k8s.io/api/core/v1"
)

type KubeNames struct {
	AppName           string
	AppURI            string
	AppImage          string
	ClusterName       string
	K8sNodeNameSuffix string
	OperatorName      string
	ServiceNames      []string
	KconfName         string
	KconfEnv          string
	DeploymentType    string
}

func GetKconfName(clusterInst *edgeproto.ClusterInst) string {
	return fmt.Sprintf("%s.%s.kubeconfig",
		clusterInst.Key.ClusterKey.Name,
		clusterInst.Key.CloudletKey.OperatorKey.Name)
}

func GetK8sNodeNameSuffix(clusterInstKey *edgeproto.ClusterInstKey) string {
	cloudletName := clusterInstKey.CloudletKey.Name
	clusterName := clusterInstKey.ClusterKey.Name
	devName := clusterInstKey.Developer
	if devName != "" {
		return NormalizeName(cloudletName + "-" + clusterName + "-" + devName)
	}
	return NormalizeName(cloudletName + "-" + clusterName)

}

func NormalizeName(name string) string {
	return util.K8SSanitize(name)
}

// GetKubeNames udpates kubeNames with normalized strings for the included clusterinst, app, and appisnt
func GetKubeNames(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*KubeNames, error) {
	if clusterInst == nil {
		return nil, fmt.Errorf("nil cluster inst")
	}
	if app == nil {
		return nil, fmt.Errorf("nil app")
	}
	if appInst == nil {
		return nil, fmt.Errorf("nil app inst")
	}
	kubeNames := KubeNames{}
	kubeNames.ClusterName = NormalizeName(clusterInst.Key.ClusterKey.Name + clusterInst.Key.Developer)
	kubeNames.K8sNodeNameSuffix = GetK8sNodeNameSuffix(&clusterInst.Key)
	kubeNames.AppName = NormalizeName(app.Key.Name)
	kubeNames.AppURI = appInst.Uri
	kubeNames.AppImage = NormalizeName(app.ImagePath)
	kubeNames.OperatorName = NormalizeName(clusterInst.Key.CloudletKey.OperatorKey.Name)
	kubeNames.KconfName = GetKconfName(clusterInst)
	kubeNames.KconfEnv = "KUBECONFIG=" + kubeNames.KconfName
	kubeNames.DeploymentType = app.Deployment
	//get service names from the yaml
	if app.Deployment == cloudcommon.AppDeploymentTypeKubernetes {
		objs, _, err := cloudcommon.DecodeK8SYaml(app.DeploymentManifest)
		if err != nil {
			return nil, fmt.Errorf("invalid kubernetes deployment yaml, %s", err.Error())
		}
		for _, o := range objs {
			log.DebugLog(log.DebugLevelMexos, "k8s obj", "obj", o)
			ksvc, ok := o.(*v1.Service)
			if !ok {
				continue
			}
			svcName := ksvc.ObjectMeta.Name
			kubeNames.ServiceNames = append(kubeNames.ServiceNames, svcName)
		}
	} else if app.Deployment == cloudcommon.AppDeploymentTypeHelm {
		// for helm chart just make sure it's the same prefix
		kubeNames.ServiceNames = append(kubeNames.ServiceNames, kubeNames.AppName)
	} else if app.Deployment == cloudcommon.AppDeploymentTypeDocker {
		// for docker use the app name
		kubeNames.ServiceNames = append(kubeNames.ServiceNames, kubeNames.AppName)
	}
	return &kubeNames, nil
}

func (k *KubeNames) ContainsService(svc string) bool {
	for _, s := range k.ServiceNames {
		if strings.HasPrefix(svc, s) {
			return true
		}
	}
	return false
}
