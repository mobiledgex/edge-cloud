package node

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var vaultRole = os.Getenv("VAULT_ROLE_ID")
var vaultSecret = os.Getenv("VAULT_SECRET_ID")
var VerifyDelay time.Duration = time.Second
var VerifyRetry = 30

type AccessKeyVerifyOnly bool

const (
	AccessKeyVerify  AccessKeyVerifyOnly = true
	AccessKeyUpgrade                     = false
)

type KeyType int

const (
	PrimaryKey = 0
	BackupKey  = 1
)

// AccessKeyClient maintains information needed on the client.
type AccessKeyClient struct {
	AccessKeyFile     string
	AccessApiAddr     string
	accessPrivKey     ed25519.PrivateKey
	cloudletKey       edgeproto.CloudletKey
	cloudletKeyStr    string
	enabled           bool
	TestSkipTlsVerify bool
	requireAccessKey  bool
}

func (s *AccessKeyClient) InitFlags() {
	flag.StringVar(&s.AccessKeyFile, "accessKeyFile", "/root/accesskey/priv.key", "access private key file")
	flag.StringVar(&s.AccessApiAddr, "accessApiAddr", "127.0.0.1:41001", "Controller's access API address")
	flag.BoolVar(&s.requireAccessKey, "requireAccessKey", true, "Require access key for RegionalCloudlet service")
}

func (s *AccessKeyClient) init(ctx context.Context, nodeType, tlsClientIssuer string, key edgeproto.CloudletKey, deploymentTag string) error {
	log.SpanLog(ctx, log.DebugLevelInfo, "access key client init")
	if tlsClientIssuer == NoTlsClientIssuer {
		// unit test mode
		log.SpanLog(ctx, log.DebugLevelInfo, "no issuer, unit-test mode")
		return nil
	}
	if tlsClientIssuer != CertIssuerRegionalCloudlet {
		// Not running on a cloudlet, no access key required.
		log.SpanLog(ctx, log.DebugLevelInfo, "not cloudlet service, no access key required")
		return nil
	}
	if s.AccessKeyFile == "" {
		return fmt.Errorf("access key not specified for cloudlet service")
	}
	if s.AccessApiAddr == "" {
		return fmt.Errorf("Controller access API address not specified")
	}
	if e2e := os.Getenv("E2ETEST_TLS"); e2e != "" {
		s.TestSkipTlsVerify = true
	}
	if deploymentTag == "dev" || deploymentTag == "stage" || deploymentTag == "upg" {
		// test setup deployment, skip cert validation
		s.TestSkipTlsVerify = true
	}
	if s.TestSkipTlsVerify {
		// for e2e and unit testing only
		log.SpanLog(ctx, log.DebugLevelInfo, "skip TLS verification testing mode")
	}
	// CloudletKey is required when using access key
	if err := key.ValidateKey(); err != nil {
		return fmt.Errorf("error access key client CloudletKey: %s", err)
	}
	keystr, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal CloudletKey: %s", err)
	}
	s.cloudletKey = key
	s.cloudletKeyStr = string(keystr)

	if nodeType == NodeTypeCRM {
		// Attempt to upgrade access key. May not exist if upgrading
		// old crm, so ignore and log error. Shepherd/DME will not
		// go through upgrade process. If correct key is backup key,
		// CRM should restore it to the primary key file for Shepherd/DME
		// to pick up.
		err = s.loadAccessKey(ctx, s.AccessKeyFile)
		log.SpanLog(ctx, log.DebugLevelInfo, "access key upgrade load", "err", err)
		// Upgrade access key
		_, err = s.upgradeAccessKey(ctx, AccessKeyUpgrade, PrimaryKey)
		if err != nil {
			// attempt to upgrade using backup key
			log.SpanLog(ctx, log.DebugLevelInfo, "upgrade failed, try backup key", "err", err)
			bkerr := s.loadAccessKey(ctx, s.backupKeyFile())
			log.SpanLog(ctx, log.DebugLevelInfo, "backup key load", "err", bkerr)
			if bkerr == nil {
				upgraded, err := s.upgradeAccessKey(ctx, AccessKeyUpgrade, BackupKey)
				if err == nil && !upgraded {
					// backup key is valid and key was not
					// upgraded, move backup to primary.
					log.SpanLog(ctx, log.DebugLevelInfo, "restore backup key")
					err = os.Rename(s.backupKeyFile(), s.AccessKeyFile)
				}
			}
		}
	} else {
		// DME/Shepherd share access key, but it may take time for
		// CRM to upgrade the access key. So retry until verified.
		// Verify ensures the access key does not require upgrade,
		// and thus will not be changed by the CRM doing upgrade.
		for ii := 0; ii < VerifyRetry; ii++ {
			if ii != 0 {
				time.Sleep(VerifyDelay)
			}
			log.SpanLog(ctx, log.DebugLevelInfo, "verify access key", "try", ii)
			err = s.loadAccessKey(ctx, s.AccessKeyFile)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "verify access key load", "err", err)
				continue
			}
			_, err = s.upgradeAccessKey(ctx, AccessKeyVerify, PrimaryKey)
			if err == nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "access key verified", "err", err)
				break
			}
			log.SpanLog(ctx, log.DebugLevelInfo, "verify access key failed", "err", err)
		}
	}
	if err != nil {
		return err
	}

	// Load access key (must succeed)
	log.SpanLog(ctx, log.DebugLevelInfo, "access key load")
	err = s.loadAccessKey(ctx, s.AccessKeyFile)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "access key client enabled")
	s.enabled = true
	return nil
}

func (s *AccessKeyClient) IsEnabled() bool {
	return s.enabled
}

func (s *AccessKeyClient) backupKeyFile() string {
	return s.AccessKeyFile + ".backup"
}

func (s *AccessKeyClient) loadAccessKey(ctx context.Context, keyFile string) error {
	log.SpanLog(ctx, log.DebugLevelInfo, "load access private key", "file", keyFile)
	// read access private key
	dat, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return err
	}
	s.accessPrivKey, err = LoadPrivPEM(dat)
	if err != nil {
		return err
	}
	return nil
}

func (s *AccessKeyClient) upgradeAccessKey(ctx context.Context, verifyOnly AccessKeyVerifyOnly, keyType KeyType) (bool, error) {
	// Request an updated AccessKey from the controller.
	// The server (controller) determines whether or not to issue a new one.
	// There are two cases where this is needed.
	//
	// 1). Upgrade from existing CRM without access key. In this case,
	// both Controller and CRM must have valid CRM Vault role/secret ids to
	// authenticate the CRM.
	//
	// 2). Upgrade from a one-time access key. One time-access keys are
	// put in heat stacks or other orchestration configs, and can only
	// be used once to upgrade to a normal access key.
	log.SpanLog(ctx, log.DebugLevelInfo, "upgradeAccessKey")
	if len(s.accessPrivKey) > 0 {
		log.SpanLog(ctx, log.DebugLevelInfo, "use access key creds")
		ctx = s.AddAccessKeySig(ctx)
	} else if vaultRole != "" && vaultSecret != "" {
		log.SpanLog(ctx, log.DebugLevelInfo, "use vault creds")
		kvPairs := []string{
			cloudcommon.AccessKeyData, s.cloudletKeyStr,
			cloudcommon.VaultKeySig, vaultRole + vaultSecret,
		}
		ctx = metadata.AppendToOutgoingContext(ctx, kvPairs...)
	} else {
		log.SpanLog(ctx, log.DebugLevelInfo, "no creds found")
		return false, fmt.Errorf("no credentials found")
	}

	tlsConfig := &tls.Config{}
	if s.TestSkipTlsVerify {
		// for e2e and unit testing only
		tlsConfig.InsecureSkipVerify = true
	}
	dialOpt := grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	clientConn, err := grpc.Dial(s.AccessApiAddr, dialOpt)
	if err != nil {
		return false, err
	}
	defer clientConn.Close()

	client := edgeproto.NewCloudletAccessKeyApiClient(clientConn)
	stream, err := client.UpgradeAccessKey(ctx)
	if err != nil {
		return false, err
	}
	// For non-CRMs that share the access key, only verify, do not upgrade.
	err = stream.Send(&edgeproto.UpgradeAccessKeyClientMsg{
		Msg:        "verify-only",
		VerifyOnly: bool(verifyOnly),
	})
	if err != nil {
		return false, err
	}

	reply, err := stream.Recv()
	if err != nil {
		return false, err
	}
	if verifyOnly {
		if reply.CrmPrivateAccessKey == "" {
			log.SpanLog(ctx, log.DebugLevelInfo, "access key verified")
			return false, nil
		}
		// should never get here, server should have sent error if
		// verification failed
		log.SpanLog(ctx, log.DebugLevelInfo, "verifyOnly unexpected response from server", "msg", reply.Msg)
		return false, fmt.Errorf("verify-only unexpected response")
	}

	if reply.CrmPrivateAccessKey == "" {
		// no upgrade required, we're done
		log.SpanLog(ctx, log.DebugLevelInfo, "no upgrade required")
		return false, nil
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "upgrading access key")
	// New key was issued. First we back up the existing key.
	// If the controller doesn't get our ack and doesn't commit
	// the new key, we can recover using the backup key.
	// Skip this if we're already using the backup key.
	backedUp := false
	if _, err := os.Stat(s.AccessKeyFile); err == nil && keyType == PrimaryKey {
		// key file exists
		log.SpanLog(ctx, log.DebugLevelInfo, "backing up existing key")
		err = os.Rename(s.AccessKeyFile, s.backupKeyFile())
		if err != nil {
			return false, err
		}
		backedUp = true
	}
	// write new key file
	log.SpanLog(ctx, log.DebugLevelInfo, "writing new key")
	err = ioutil.WriteFile(s.AccessKeyFile, []byte(reply.CrmPrivateAccessKey), 0600)
	if err != nil {
		if backedUp {
			// undo changes
			undoErr := os.Rename(s.backupKeyFile(), s.AccessKeyFile)
			log.SpanLog(ctx, log.DebugLevelApi, "restore from backup", "err", undoErr)
		}
		return false, err
	}
	// We now have the new key on disk, plus the old key. Ack the changes.
	// If connection fails at this point, Controller may or may not commit
	// the new key to etcd. If it doesn't, we can authenticate using the
	// backup key. If it does, we will authenticate using the regular key.
	// So no special handling is needed on failure.
	log.SpanLog(ctx, log.DebugLevelInfo, "sending ack")
	err = stream.Send(&edgeproto.UpgradeAccessKeyClientMsg{
		Msg: "ack-new-key",
	})
	if err != nil {
		return true, err
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "upgradeAccessKey complete")
	// wait for commit-complete message, otherwise we may use new key
	// before Controller has updated its caches.
	reply, err = stream.Recv()
	if err != nil {
		return true, err
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "recv commit-complete", "reply", reply)
	return true, nil
}

// Add an access key signature to the grpc metadata
func (s *AccessKeyClient) AddAccessKeySig(ctx context.Context) context.Context {
	sig := ed25519.Sign(s.accessPrivKey, []byte(s.cloudletKeyStr))
	sigb64 := base64.StdEncoding.EncodeToString(sig)

	kvPairs := []string{
		cloudcommon.AccessKeyData, s.cloudletKeyStr,
		cloudcommon.AccessKeySig, sigb64,
	}
	log.SpanLog(ctx, log.DebugLevelApi, "adding access key to signature", "sig", sigb64)
	return metadata.AppendToOutgoingContext(ctx, kvPairs...)
}

// Grpc unary interceptor to add access key
func (s *AccessKeyClient) UnaryAddAccessKey(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if s.enabled {
		ctx = s.AddAccessKeySig(ctx)
	}
	return invoker(ctx, method, req, resp, cc, opts...)
}

// Grpc stream interceptor to add access key
func (s *AccessKeyClient) StreamAddAccessKey(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if s.enabled {
		ctx = s.AddAccessKeySig(ctx)
	}
	return streamer(ctx, desc, cc, method, opts...)
}

// Common helper function to connect to Controller
func (s *AccessKeyClient) ConnectController(ctx context.Context) (*grpc.ClientConn, error) {
	var tlsConfig *tls.Config
	// TLS access config for talking to the Controller's AccessApi endpoint.
	// The Controller will have a letsencrypt-public issued certificate, and will
	// not require a client certificate.
	skipVerify := false
	if s.TestSkipTlsVerify {
		// unit/e2e-tests use fake cert, so skip verification
		skipVerify = true
	}
	tlsConfig = &tls.Config{
		ServerName:         strings.Split(s.AccessApiAddr, ":")[0],
		InsecureSkipVerify: skipVerify,
	}
	dialOption := edgetls.GetGrpcDialOption(tlsConfig)
	// sign request with access key
	return grpc.Dial(s.AccessApiAddr, dialOption, grpc.WithBlock(),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			log.UnaryClientTraceGrpc,
			s.UnaryAddAccessKey,
		)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			log.StreamClientTraceGrpc,
			s.StreamAddAccessKey,
		)),
	)
}
