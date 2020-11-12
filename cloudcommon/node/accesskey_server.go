package node

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// Store and retrieve verified CloudletKey on context

var BadAuthDelay = 3 * time.Second
var UpgradeAccessKeyMethod = "/edgeproto.CloudletAccessKeyApi/UpgradeAccessKey"

type accessKeyVerifiedTagType string

const accessKeyVerifiedTag accessKeyVerifiedTagType = "accessKeyVerified"

type AccessKeyVerified struct {
	Key             edgeproto.CloudletKey
	UpgradeRequired bool
}

func ContextSetAccessKeyVerified(ctx context.Context, info *AccessKeyVerified) context.Context {
	return context.WithValue(ctx, accessKeyVerifiedTag, info)
}

func ContextGetAccessKeyVerified(ctx context.Context) *AccessKeyVerified {
	key, ok := ctx.Value(accessKeyVerifiedTag).(*AccessKeyVerified)
	if !ok {
		return nil
	}
	return key
}

// AccessKeyServer maintains state to validate clients.
type AccessKeyServer struct {
	cloudletCache       *edgeproto.CloudletCache
	crmVaultRole        string
	crmVaultSecret      string
	requireTlsAccessKey bool
}

func NewAccessKeyServer(cloudletCache *edgeproto.CloudletCache) *AccessKeyServer {
	server := &AccessKeyServer{
		cloudletCache:  cloudletCache,
		crmVaultRole:   os.Getenv("CRM_VAULT_ROLE_ID"),
		crmVaultSecret: os.Getenv("CRM_VAULT_SECRET_ID"),
	}
	// for testing, reduce bad auth delay
	if e2e := os.Getenv("E2ETEST_TLS"); e2e != "" {
		BadAuthDelay = time.Millisecond
	}
	return server
}

func (s *AccessKeyServer) SetCrmVaultAuth(role, secret string) {
	s.crmVaultRole = role
	s.crmVaultSecret = secret
}

func (s *AccessKeyServer) SetRequireTlsAccessKey(require bool) {
	s.requireTlsAccessKey = require
}

// Verify an access key signature in the grpc metadata
func (s *AccessKeyServer) VerifyAccessKeySig(ctx context.Context, method string) (*AccessKeyVerified, error) {
	// grab CloudletKey and signature from grpc metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no meta data on grpc context")
	}
	data, found := md[cloudcommon.AccessKeyData]
	if !found || len(data) == 0 {
		return nil, fmt.Errorf("error, %s not found in metadata", cloudcommon.AccessKeyData)
	}
	verified := &AccessKeyVerified{}

	// data is the cloudlet key
	err := json.Unmarshal([]byte(data[0]), &verified.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cloudlet key from metadata, %s, %s", data, err)
	}
	// find public key to validate signature
	cloudlet := edgeproto.Cloudlet{}
	if !s.cloudletCache.Get(&verified.Key, &cloudlet) {
		return nil, fmt.Errorf("failed to find cloudlet %s to verify access key", data)
	}

	// look up key signature
	sigb64, found := md[cloudcommon.AccessKeySig]
	if found && len(sigb64) > 0 {
		// access key signature
		sig, err := base64.StdEncoding.DecodeString(sigb64[0])
		if err != nil {
			return nil, fmt.Errorf("failed to base64 decode access key signature, %v", err)
		}
		if cloudlet.CrmAccessPublicKey == "" {
			return nil, fmt.Errorf("No crm access public key registered for cloudlet %s", data)
		}
		if cloudlet.CrmAccessKeyUpgradeRequired && method != UpgradeAccessKeyMethod {
			return nil, fmt.Errorf("access key requires upgrade, does not allow api call %s", method)
		}
		verified.UpgradeRequired = cloudlet.CrmAccessKeyUpgradeRequired

		// public key is saved as PEM
		pubKey, err := LoadPubPEM([]byte(cloudlet.CrmAccessPublicKey))
		if err != nil {
			return nil, fmt.Errorf("Failed to decode crm public access key, %s, %s", data, err)
		}
		ok = ed25519.Verify(pubKey, []byte(data[0]), sig)
		if !ok {
			return nil, fmt.Errorf("failed to verify cloudlet access key signature")
		}
		log.SpanLog(ctx, log.DebugLevelApi, "verified access key", "CloudletKey", verified.Key)
		return verified, nil
	}
	vaultSig, found := md[cloudcommon.VaultKeySig]
	if found && len(vaultSig) > 0 {
		// vault key signature - only allowed for UpgradeAccessKey
		if method != UpgradeAccessKeyMethod {
			return nil, fmt.Errorf("vault auth not allowed for api %s", method)
		}
		verified.UpgradeRequired = true

		crmVaultKey := s.crmVaultRole + s.crmVaultSecret
		if crmVaultKey == "" {
			// Controller is not configured to allow Vault-based auth
			// for backwards compatibility.
			return nil, fmt.Errorf("Vault-based auth not allowed")
		}
		if crmVaultKey != vaultSig[0] {
			return nil, fmt.Errorf("Vault-based auth key mismatch")
		}
		log.SpanLog(ctx, log.DebugLevelApi, "verified vault keys", "CloudletKey", verified.Key)
		return verified, nil
	}
	return nil, fmt.Errorf("no valid auth found")
}

// Grpc unary interceptor to require and verify access key
func (s *AccessKeyServer) UnaryRequireAccessKey(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "unary requiring access key")
	verified, err := s.VerifyAccessKeySig(ctx, info.FullMethod)
	if err != nil {
		// We intentionally do not return detailed errors, to avoid leaking of
		// information to malicious attackers, much like a usual "login"
		// function behaves.
		log.SpanLog(ctx, log.DebugLevelApi, "accesskey auth failed", "err", err)
		time.Sleep(BadAuthDelay)
		return nil, err
	}
	ctx = ContextSetAccessKeyVerified(ctx, verified)
	return handler(ctx, req)
}

// Grpc stream interceptor to require and verify access key
func (s *AccessKeyServer) StreamRequireAccessKey(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "stream requiring access key")
	verified, err := s.VerifyAccessKeySig(ctx, info.FullMethod)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "accesskey auth failed", "err", err)
		time.Sleep(BadAuthDelay)
		return err
	}
	ctx = ContextSetAccessKeyVerified(ctx, verified)
	// override context on existing stream, since no way to set it
	stream = cloudcommon.WrapStream(stream, ctx)
	return handler(srv, stream)

}

// Grpc unary interceptor to require and verify access key based on client cert
func (s *AccessKeyServer) UnaryTlsAccessKey(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	required, err := s.isTlsAccessKeyRequired(ctx)
	if err != nil {
		return nil, err
	}
	if required {
		return s.UnaryRequireAccessKey(ctx, req, info, handler)
	}
	return handler(ctx, req)
}

// Grpc stream interceptor to require and verify access key based on client cert
func (s *AccessKeyServer) StreamTlsAccessKey(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	required, err := s.isTlsAccessKeyRequired(stream.Context())
	if err != nil {
		return err
	}
	if required {
		return s.StreamRequireAccessKey(srv, stream, info, handler)
	}
	return handler(srv, stream)
}

// Determines from the grpc context if an access key is required.
func (s *AccessKeyServer) isTlsAccessKeyRequired(ctx context.Context) (bool, error) {
	if !s.requireTlsAccessKey {
		return false, nil
	}
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return false, fmt.Errorf("no grpc peer context")
	}
	tlsInfo, ok := pr.AuthInfo.(credentials.TLSInfo)
	if ok {
		for _, chain := range tlsInfo.State.VerifiedChains {
			for _, cert := range chain {
				if !cert.IsCA || len(cert.DNSNames) == 0 {
					continue
				}
				commonName := cert.DNSNames[0]
				// if cert is issued by regional-access-key,
				// then access key verification is required.
				if commonName == CertIssuerRegionalCloudlet {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (s *AccessKeyServer) UpgradeAccessKey(stream edgeproto.CloudletAccessKeyApi_UpgradeAccessKeyServer, commitKeyFunc func(ctx context.Context, key *edgeproto.CloudletKey, pubPEM string) error) error {
	ctx := stream.Context()
	verified := ContextGetAccessKeyVerified(ctx)
	if verified == nil {
		// this should never happen, the interceptor should error out first
		return fmt.Errorf("access key not verified")
	}
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	if msg.VerifyOnly {
		log.SpanLog(ctx, log.DebugLevelApi, "access key verifyOnly")
		// Non-CRM service that is verifying the access key.
		// Fail verification if upgrade is required.
		if verified.UpgradeRequired {
			return fmt.Errorf("access key upgrade required")
		}
		return stream.Send(&edgeproto.UpgradeAccessKeyServerMsg{
			Msg: "verified",
		})
	}

	if !verified.UpgradeRequired {
		log.SpanLog(ctx, log.DebugLevelApi, "access key upgrade not required")
		return stream.Send(&edgeproto.UpgradeAccessKeyServerMsg{
			Msg: "upgrade-not-needed",
		})
	}
	log.SpanLog(ctx, log.DebugLevelApi, "generating new access key")
	// upgrade required, generate new key
	keyPair, err := GenerateAccessKey()
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelApi, "sending new access key")
	err = stream.Send(&edgeproto.UpgradeAccessKeyServerMsg{
		Msg:                 "new-key",
		CrmPrivateAccessKey: keyPair.PrivatePEM,
	})
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelApi, "waiting for ack")
	// Read ack to make sure CRM got new key.
	// See comments in client code for UpgradeAccessKey for error recovery.
	_, err = stream.Recv()
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelApi, "ack received, committing new key")
	err = commitKeyFunc(ctx, &verified.Key, keyPair.PublicPEM)
	if err != nil {
		return err
	}
	// this final send makes client wait until new public key pem
	// has been committed, otherwise client will try to connect
	// immediately with new private key before new public key has
	// been put into cache by etcd watch.
	return stream.Send(&edgeproto.UpgradeAccessKeyServerMsg{
		Msg: "commit-complete",
	})
}

// AccessKeyGrcpServer starts up a grpc listener for the access API endpoint.
// This is used both by the Controller and various unit test code, and keeps
// the interceptor setup consistent while avoiding duplicate code.
type AccessKeyGrpcServer struct {
	lis             net.Listener
	serv            *grpc.Server
	AccessKeyServer *AccessKeyServer
}

func (s *AccessKeyGrpcServer) Start(addr string, keyServer *AccessKeyServer, tlsConfig *tls.Config, registerHandlers func(server *grpc.Server)) error {
	s.AccessKeyServer = keyServer

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.lis = lis

	// start AccessKey grpc service.
	grpcServer := grpc.NewServer(cloudcommon.GrpcCreds(tlsConfig),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			cloudcommon.AuditUnaryInterceptor,
			s.AccessKeyServer.UnaryRequireAccessKey,
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			cloudcommon.AuditStreamInterceptor,
			s.AccessKeyServer.StreamRequireAccessKey,
		)),
	)
	if registerHandlers != nil {
		registerHandlers(grpcServer)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.FatalLog("Failed to serve", "err", err)
		}
	}()
	s.serv = grpcServer
	return nil
}

func (s *AccessKeyGrpcServer) Stop() {
	if s.serv != nil {
		s.serv.Stop()
		s.serv = nil
	}
	if s.lis != nil {
		s.lis.Close()
		s.lis = nil
	}
}

func (s *AccessKeyGrpcServer) ApiAddr() string {
	return s.lis.Addr().String()
}

// Basic edgeproto.CloudletAccessKeyApiServer handler for unit tests only.
type BasicUpgradeHandler struct {
	KeyServer *AccessKeyServer
}

func (s *BasicUpgradeHandler) UpgradeAccessKey(stream edgeproto.CloudletAccessKeyApi_UpgradeAccessKeyServer) error {
	return s.KeyServer.UpgradeAccessKey(stream, s.commitKey)
}

func (s *BasicUpgradeHandler) commitKey(ctx context.Context, key *edgeproto.CloudletKey, pubPEM string) error {
	// Not thread safe, unit-test only.
	cloudlet := &edgeproto.Cloudlet{}
	if !s.KeyServer.cloudletCache.Get(key, cloudlet) {
		return key.NotFoundError()
	}
	cloudlet.CrmAccessPublicKey = pubPEM
	cloudlet.CrmAccessKeyUpgradeRequired = false
	s.KeyServer.cloudletCache.Update(ctx, cloudlet, 0)
	return nil
}
