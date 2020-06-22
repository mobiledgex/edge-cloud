package cloudcommon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

func GetCloudletLogFile(filePrefix string) string {
	return "/tmp/" + filePrefix + ".log"
}

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
	testMode := false
	span := ""
	cloudletVMImagePath := ""
	region := ""
	commercialCerts := false
	useVaultCAs := false
	useVaultCerts := false
	appDNSRoot := ""
	chefServerPath := ""
	deploymentTag := ""
	if pfConfig != nil {
		for k, v := range pfConfig.EnvVar {
			envVars[k] = v
		}
		notifyCtrlAddrs = pfConfig.NotifyCtrlAddrs
		tlsCertFile = pfConfig.TlsCertFile
		vaultAddr = pfConfig.VaultAddr
		testMode = pfConfig.TestMode
		span = pfConfig.Span
		cloudletVMImagePath = pfConfig.CloudletVmImagePath
		region = pfConfig.Region
		commercialCerts = pfConfig.CommercialCerts
		useVaultCAs = pfConfig.UseVaultCas
		useVaultCerts = pfConfig.UseVaultCerts
		appDNSRoot = pfConfig.AppDnsRoot
		chefServerPath = pfConfig.ChefServerPath
		deploymentTag = pfConfig.DeploymentTag
	}
	for envKey, envVal := range cloudlet.EnvVar {
		envVars[envKey] = envVal
	}

	opts = append(opts, process.WithDebug("api,infra,notify,info"))

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
		VaultAddr:           vaultAddr,
		PhysicalName:        cloudlet.PhysicalName,
		TestMode:            testMode,
		Span:                span,
		ContainerVersion:    cloudlet.ContainerVersion,
		VMImageVersion:      cloudlet.VmImageVersion,
		CloudletVMImagePath: cloudletVMImagePath,
		Region:              region,
		CommercialCerts:     commercialCerts,
		UseVaultCAs:         useVaultCAs,
		UseVaultCerts:       useVaultCerts,
		AppDNSRoot:          appDNSRoot,
		ChefServerPath:      chefServerPath,
		DeploymentTag:       deploymentTag,
	}, opts, nil
}

func GetCRMCmd(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (string, *map[string]string, error) {
	crmProc, opts, err := getCrmProc(cloudlet, pfConfig)
	if err != nil {
		return "", nil, err
	}

	return crmProc.String(opts...), &crmProc.Common.EnvVars, nil
}

func GetCRMCmdArgs(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) ([]string, *map[string]string, error) {
	crmProc, opts, err := getCrmProc(cloudlet, pfConfig)
	if err != nil {
		return nil, nil, err
	}

	return crmProc.GetArgs(opts...), &crmProc.Common.EnvVars, nil
}

var trackedProcess = map[edgeproto.CloudletKey]*process.Crm{}

func StartCRMService(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) error {
	log.SpanLog(ctx, log.DebugLevelApi, "start crmserver", "cloudlet", cloudlet.Key)

	// Get non-conflicting port for NotifySrvAddr if actual port is 0
	newAddr, err := GetAvailablePort(cloudlet.NotifySrvAddr)
	if err != nil {
		return err
	}
	cloudlet.NotifySrvAddr = newAddr

	trackedProcess[cloudlet.Key] = nil
	crmProc, opts, err := getCrmProc(cloudlet, pfConfig)
	if err != nil {
		return err
	}

	err = crmProc.StartLocal(GetCloudletLogFile(cloudlet.Key.Name), opts...)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelApi, "started "+crmProc.GetExeName(), "pfConfig", pfConfig)
	trackedProcess[cloudlet.Key] = crmProc

	return nil
}

// StopCRMService stops the crmserver on the specified cloudlet, or kills any
// crm process if the cloudlet specified is nil
func StopCRMService(ctx context.Context, cloudlet *edgeproto.Cloudlet) error {
	args := ""
	if cloudlet != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "stop crmserver", "cloudlet", cloudlet.Key)
		crmProc, _, err := getCrmProc(cloudlet, nil)
		if err != nil {
			return err
		}
		args = util.EscapeJson(crmProc.LookupArgs())
	}

	// max wait time for process to go down gracefully, after which it is killed forcefully
	maxwait := 10 * time.Millisecond

	c := make(chan string)
	go process.KillProcessesByName("crmserver", maxwait, args, c)

	log.SpanLog(ctx, log.DebugLevelInfra, "stopped crmserver", "msg", <-c)

	// After above, processes will be in Zombie state. Hence use wait to cleanup the processes
	if cloudlet != nil {
		if cmdProc, ok := trackedProcess[cloudlet.Key]; ok {
			// Wait is in a goroutine as it is blocking call if
			// process is not killed for some reasons
			go cmdProc.Wait()
			delete(trackedProcess, cloudlet.Key)
		}
	} else {
		for _, v := range trackedProcess {
			go v.Wait()
		}
		trackedProcess = make(map[edgeproto.CloudletKey]*process.Crm)
	}
	return nil
}

// Parses cloudlet logfile and fetches FatalLog output
func GetCloudletLog(ctx context.Context, key *edgeproto.CloudletKey) (string, error) {
	logFile := GetCloudletLogFile(key.Name)
	log.SpanLog(ctx, log.DebugLevelApi, fmt.Sprintf("parse cloudlet logfile %s to fetch crash details", logFile))

	file, err := os.Open(logFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	out := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "FATAL") {
			fields := strings.Fields(line)
			if len(fields) > 3 {
				out = strings.Join(fields[3:], " ")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return out, nil
}

func CrmServiceWait(key edgeproto.CloudletKey) error {
	if _, ok := trackedProcess[key]; ok {
		err := trackedProcess[key].Wait()
		if err != nil && strings.Contains(err.Error(), "Wait was already called") {
			return nil
		}
		delete(trackedProcess, key)
		return fmt.Errorf("Crm Service Stopped: %v", err)
	}
	return nil
}
