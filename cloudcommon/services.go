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
	if pfConfig != nil {
		envVars["VAULT_ROLE_ID"] = pfConfig.CrmRoleId
		envVars["VAULT_SECRET_ID"] = pfConfig.CrmSecretId
		notifyCtrlAddrs = pfConfig.NotifyCtrlAddrs
		tlsCertFile = pfConfig.TlsCertFile
		vaultAddr = pfConfig.VaultAddr
		testMode = pfConfig.TestMode
		span = pfConfig.Span
	}
	for envKey, envVal := range cloudlet.EnvVars {
		envVars[envKey] = envVal
	}

	opts = append(opts, process.WithDebug("api,mexos,notify,info"))

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
		TestMode:     testMode,
		Span:         span,
	}, opts, nil
}

func GetCRMCmd(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (string, *map[string]string, error) {
	crmProc, opts, err := getCrmProc(cloudlet, pfConfig)
	if err != nil {
		return "", nil, err
	}

	return crmProc.String(opts...), &crmProc.Common.EnvVars, nil
}

var trackedProcess = map[edgeproto.CloudletKey]*process.Crm{}

func StartCRMService(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) error {
	log.SpanLog(ctx, log.DebugLevelApi, "start crmserver", "cloudlet", cloudlet.Key)

	trackedProcess[cloudlet.Key] = nil
	crmProc, opts, err := getCrmProc(cloudlet, pfConfig)
	if err != nil {
		return err
	}

	err = crmProc.StartLocal(GetCloudletLogFile(cloudlet.Key.Name), opts...)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelApi, "started "+crmProc.GetExeName())
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
		args = crmProc.LookupArgs()
	}

	// max wait time for process to go down gracefully, after which it is killed forcefully
	maxwait := 10 * time.Millisecond

	c := make(chan string)
	go process.KillProcessesByName("crmserver", maxwait, args, c)

	log.SpanLog(ctx, log.DebugLevelMexos, "stopped crmserver", "msg", <-c)

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
		trackedProcess[key].Wait()
		delete(trackedProcess, key)
		return fmt.Errorf("Crm Service Stopped")
	}
	return nil
}
