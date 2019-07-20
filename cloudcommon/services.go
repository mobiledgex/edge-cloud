package cloudcommon

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
)

func getCrmProc(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (*process.Crm, []process.StartOp, error) {
	opts := []process.StartOp{}

	cloudletKeyStr, err := json.Marshal(cloudlet.Key)
	if err != nil {
		return nil, opts, fmt.Errorf("unable to marshal cloudlet key")
	}

	envVars := make(map[string]string)
	notifyCtrlAddrs := ""
	tlsCertFile := ""
	vaultAddr := ""
	if pfConfig != nil {
		envVars["VAULT_ROLE_ID"] = pfConfig.CrmRoleId
		envVars["VAULT_SECRET_ID"] = pfConfig.CrmSecretId
		notifyCtrlAddrs = pfConfig.NotifyCtrlAddrs
		tlsCertFile = pfConfig.TlsCertFile
		vaultAddr = pfConfig.VaultAddr
	}

	opts = append(opts, process.WithDebug("api,mexos,notify"))

	return &process.Crm{
		NotifyAddrs:   notifyCtrlAddrs,
		NotifySrvAddr: cloudlet.NotifySrvAddr,
		CloudletKey:   string(cloudletKeyStr),
		Platform:      cloudlet.PlatformType.String(),
		Common: process.Common{
			Hostname: cloudlet.Key.Name,
			EnvVars:  envVars,
		},
		TLS: process.TLSCerts{
			ServerCert: tlsCertFile,
		},
		VaultAddr:    vaultAddr,
		PhysicalName: cloudlet.PhysicalName,
	}, opts, nil
}

func GetCRMCmd(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (string, *map[string]string, error) {
	crmProc, opts, err := getCrmProc(cloudlet, pfConfig)
	if err != nil {
		return "", nil, err
	}

	return crmProc.String(opts...), &crmProc.Common.EnvVars, nil
}

func StartCRMService(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) error {
	crmProc, opts, err := getCrmProc(cloudlet, pfConfig)
	if err != nil {
		return err
	}

	err = crmProc.StartLocal("/tmp/"+cloudlet.Key.Name+".log", opts...)
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "started "+crmProc.GetExeName())

	return nil
}

func StopCRMService(cloudlet *edgeproto.Cloudlet) error {
	crmProc, _, err := getCrmProc(cloudlet, nil)
	if err != nil {
		return err
	}
	// max wait time for process to go down gracefully, after which it is killed forcefully
	maxwait := 10 * time.Millisecond

	c := make(chan string)
	go process.KillProcessesByName(crmProc.GetExeName(), maxwait, crmProc.LookupArgs(), c)

	log.DebugLog(log.DebugLevelMexos, "stopped crmserver", "msg", <-c)
	return nil
}
