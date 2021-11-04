package cloudcommon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

func GetCloudletLogFile(filePrefix string) string {
	return "/tmp/" + filePrefix + ".log"
}

func GetLocalAccessKeyDir() string {
	return "/tmp/accesskeys"
}

func GetLocalAccessKeyFile(filePrefix string, haRole process.HARole) string {
	return GetLocalAccessKeyDir() + "/" + filePrefix + string(haRole) + ".key"
}

func GetCrmAccessKeyFile() string {
	return "/root/accesskey/accesskey.pem"
}

func getCrmProc(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, HARole process.HARole) (*process.Crm, []process.StartOp, error) {
	opts := []process.StartOp{}

	cloudletKeyStr, err := json.Marshal(cloudlet.Key)
	if err != nil {
		return nil, opts, fmt.Errorf("unable to marshal cloudlet key")
	}

	envVars := make(map[string]string)
	notifyCtrlAddrs := ""
	tlsCertFile := ""
	tlsKeyFile := ""
	tlsCAFile := ""
	vaultAddr := ""
	testMode := false
	span := ""
	cloudletVMImagePath := ""
	region := ""
	commercialCerts := false
	useVaultPki := false
	appDNSRoot := ""
	chefServerPath := ""
	deploymentTag := ""
	accessApiAddr := ""
	cacheDir := ""
	if pfConfig != nil {
		for k, v := range pfConfig.EnvVar {
			envVars[k] = v
		}
		notifyCtrlAddrs = pfConfig.NotifyCtrlAddrs
		tlsCertFile = pfConfig.TlsCertFile
		tlsKeyFile = pfConfig.TlsKeyFile
		tlsCAFile = pfConfig.TlsCaFile
		testMode = pfConfig.TestMode
		span = pfConfig.Span
		cloudletVMImagePath = pfConfig.CloudletVmImagePath
		region = pfConfig.Region
		commercialCerts = pfConfig.CommercialCerts
		useVaultPki = pfConfig.UseVaultPki
		appDNSRoot = pfConfig.AppDnsRoot
		chefServerPath = pfConfig.ChefServerPath
		deploymentTag = pfConfig.DeploymentTag
		accessApiAddr = pfConfig.AccessApiAddr
		cacheDir = pfConfig.CacheDir
	}
	for envKey, envVal := range cloudlet.EnvVar {
		envVars[envKey] = envVal
	}

	opts = append(opts, process.WithDebug("api,infra,notify,info"))

	notAddr := cloudlet.NotifySrvAddr
	if HARole == process.HARoleSecondary {
		notAddr = cloudlet.SecondaryNotifySrvAddr
	}
	return &process.Crm{
		NotifyAddrs:   notifyCtrlAddrs,
		NotifySrvAddr: notAddr,
		CloudletKey:   string(cloudletKeyStr),
		Platform:      cloudlet.PlatformType.String(),
		Common: process.Common{
			Hostname: cloudlet.Key.Name,
			EnvVars:  envVars,
		},
		NodeCommon: process.NodeCommon{
			TLS: process.TLSCerts{
				ServerCert: tlsCertFile,
				ServerKey:  tlsKeyFile,
				CACert:     tlsCAFile,
			},
			VaultAddr:     vaultAddr,
			UseVaultPki:   useVaultPki,
			DeploymentTag: deploymentTag,
			AccessApiAddr: accessApiAddr,
		},
		PhysicalName:        cloudlet.PhysicalName,
		TestMode:            testMode,
		Span:                span,
		ContainerVersion:    cloudlet.ContainerVersion,
		VMImageVersion:      cloudlet.VmImageVersion,
		CloudletVMImagePath: cloudletVMImagePath,
		Region:              region,
		CommercialCerts:     commercialCerts,
		AppDNSRoot:          appDNSRoot,
		ChefServerPath:      chefServerPath,
		CacheDir:            cacheDir,
		HARole:              HARole,
	}, opts, nil
}

type trackedProcessKey struct {
	cloudletKey edgeproto.CloudletKey
	haRole      process.HARole
}

var trackedProcess = map[trackedProcessKey]*process.Crm{}
var trackedProcessMux sync.Mutex

func GetCRMCmdArgs(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, haRole process.HARole) ([]string, *map[string]string, error) {
	crmProc, opts, err := getCrmProc(cloudlet, pfConfig, haRole)
	if err != nil {
		return nil, nil, err
	}
	crmProc.AccessKeyFile = GetCrmAccessKeyFile()

	return crmProc.GetArgs(opts...), &crmProc.Common.EnvVars, nil
}

func StartCRMService(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, haRole process.HARole) error {
	log.SpanLog(ctx, log.DebugLevelApi, "start crmserver", "cloudlet", cloudlet.Key, "haRole", haRole)

	// Get non-conflicting port for NotifySrvAddr if actual port is 0
	var newAddr string
	var err error
	if haRole == process.HARoleSecondary {
		newAddr, err = GetAvailablePort(cloudlet.SecondaryNotifySrvAddr)
		cloudlet.SecondaryNotifySrvAddr = newAddr
	} else {
		newAddr, err = GetAvailablePort(cloudlet.NotifySrvAddr)
		cloudlet.NotifySrvAddr = newAddr
	}
	if err != nil {
		return err
	}
	ak := pfConfig.CrmAccessPrivateKey
	if haRole == process.HARoleSecondary {
		ak = pfConfig.CrmSecondaryAccessPrivateKey
	}
	accessKeyFile := GetLocalAccessKeyFile(cloudlet.Key.Name, haRole)
	if ak != "" {
		// Write access key to local disk
		err = os.MkdirAll(GetLocalAccessKeyDir(), 0744)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(accessKeyFile, []byte(ak), 0644)
		if err != nil {
			return err
		}
	}

	// track all local crm processes
	procKey := trackedProcessKey{
		cloudletKey: cloudlet.Key,
		haRole:      haRole,
	}
	trackedProcessMux.Lock()
	trackedProcess[procKey] = nil
	trackedProcessMux.Unlock()
	crmProc, opts, err := getCrmProc(cloudlet, pfConfig, haRole)
	if err != nil {
		return err
	}
	crmProc.AccessKeyFile = accessKeyFile
	crmProc.HARole = haRole

	filePrefix := cloudlet.Key.Name
	if haRole != process.HARoleNone {
		filePrefix += "-" + string(haRole)
	}

	err = crmProc.StartLocal(GetCloudletLogFile(filePrefix), opts...)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelApi, "started "+crmProc.GetExeName(), "pfConfig", pfConfig)
	trackedProcessMux.Lock()
	trackedProcess[procKey] = crmProc
	trackedProcessMux.Unlock()

	return nil
}

// StopCRMService stops the crmserver on the specified cloudlet, or kills any
// crm process if the cloudlet specified is nil
func StopCRMService(ctx context.Context, cloudlet *edgeproto.Cloudlet, haRole process.HARole) error {
	args := ""
	if cloudlet != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "stop crmserver", "cloudlet", cloudlet.Key)
		crmProc, _, err := getCrmProc(cloudlet, nil, haRole)
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
	trackedProcessMux.Lock()
	if cloudlet != nil {
		procKey := trackedProcessKey{
			cloudletKey: cloudlet.Key,
			haRole:      haRole,
		}
		if cmdProc, ok := trackedProcess[procKey]; ok {
			// Wait is in a goroutine as it is blocking call if
			// process is not killed for some reasons
			go cmdProc.Wait()
			delete(trackedProcess, procKey)
		}
	} else {
		for _, v := range trackedProcess {
			go v.Wait()
		}
		trackedProcess = make(map[trackedProcessKey]*process.Crm)
	}
	trackedProcessMux.Unlock()
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

	roles := []process.HARole{
		process.HARoleNone,
		process.HARolePrimary,
		process.HARoleSecondary,
	}
	// loop through all possible HA roles to find running CRMs
	var crmProcs []*process.Crm
	trackedProcessMux.Lock()
	for _, r := range roles {
		procKey := trackedProcessKey{
			cloudletKey: key,
			haRole:      r,
		}
		tp, ok := trackedProcess[procKey]
		delete(trackedProcess, procKey)
		if ok {
			crmProcs = append(crmProcs, tp)
		}
	}
	trackedProcessMux.Unlock()
	for _, p := range crmProcs {
		err := p.Wait()
		if err != nil && strings.Contains(err.Error(), "Wait was already called") {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Crm Service Stopped: %v", err)
		}
	}
	return nil
}
