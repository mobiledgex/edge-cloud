package accessapi

import (
	"context"
	"encoding/json"

	"github.com/cloudflare/cloudflare-go"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/chefmgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/vault"
)

// AccessData types
const (
	GetCloudletAccessVars   = "get-cloudlet-access-vars"
	GetRegistryAuth         = "get-registry-auth"
	SignSSHKey              = "sign-ssh-key"
	GetSSHPublicKey         = "get-ssh-public-key"
	GetOldSSHKey            = "get-old-ssh-key"
	GetChefAuthKey          = "get-chef-auth-key"
	CreateOrUpdateDNSRecord = "create-or-update-dns-record"
	GetDNSRecords           = "get-dns-records"
	DeleteDNSRecord         = "delete-dns-record"
	GetSessionTokens        = "get-session-tokens"
	GetPublicCert           = "get-public-cert"
	GetKafkaCreds           = "get-kafka-creds"
	GetGCSCreds             = "get-gcs-creds"
)

// ControllerClient implements platform.AccessApi for cloudlet
// services by connecting to the Controller.
// To avoid having to change the Controller's API if we need to
// add new functions to the platform.AccessApi interface, all
// requests to the Controller go through a generic single API.
// Data is marshaled here. Unmarshaling is done in ControllerHandler.
type ControllerClient struct {
	client edgeproto.CloudletAccessApiClient
}

func NewControllerClient(client edgeproto.CloudletAccessApiClient) *ControllerClient {
	return &ControllerClient{
		client: client,
	}
}

func (s *ControllerClient) GetCloudletAccessVars(ctx context.Context) (map[string]string, error) {
	req := &edgeproto.AccessDataRequest{
		Type: GetCloudletAccessVars,
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return nil, err
	}
	vars := map[string]string{}
	err = json.Unmarshal(reply.Data, &vars)
	return vars, err
}

func (s *ControllerClient) GetRegistryAuth(ctx context.Context, imgUrl string) (*cloudcommon.RegistryAuth, error) {
	req := &edgeproto.AccessDataRequest{
		Type: GetRegistryAuth,
		Data: []byte(imgUrl),
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return nil, err
	}
	auth := &cloudcommon.RegistryAuth{}
	err = json.Unmarshal(reply.Data, auth)
	return auth, err
}

func (s *ControllerClient) SignSSHKey(ctx context.Context, publicKey string) (string, error) {
	req := &edgeproto.AccessDataRequest{
		Type: SignSSHKey,
		Data: []byte(publicKey),
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return "", err
	}
	return string(reply.Data), nil
}

func (s *ControllerClient) GetSSHPublicKey(ctx context.Context) (string, error) {
	req := &edgeproto.AccessDataRequest{
		Type: GetSSHPublicKey,
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return "", err
	}
	return string(reply.Data), nil
}

func (s *ControllerClient) GetOldSSHKey(ctx context.Context) (*vault.MEXKey, error) {
	req := &edgeproto.AccessDataRequest{
		Type: GetOldSSHKey,
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return nil, err
	}
	mexKey := &vault.MEXKey{}
	err = json.Unmarshal(reply.Data, mexKey)
	return mexKey, err
}

func (s *ControllerClient) GetChefAuthKey(ctx context.Context) (*chefmgmt.ChefAuthKey, error) {
	req := &edgeproto.AccessDataRequest{
		Type: GetChefAuthKey,
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return nil, err
	}
	auth := &chefmgmt.ChefAuthKey{}
	err = json.Unmarshal(reply.Data, auth)
	return auth, err
}

func (s *ControllerClient) GetPublicCert(ctx context.Context, commonName string) (*vault.PublicCert, error) {
	req := &edgeproto.AccessDataRequest{
		Type: GetPublicCert,
		Data: []byte(commonName),
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return nil, err
	}
	pubcert := &vault.PublicCert{}
	err = json.Unmarshal(reply.Data, pubcert)
	return pubcert, err
}

type DNSRequest struct {
	Zone    string
	Name    string
	RType   string
	Content string
	TTL     int
	Proxy   bool
}

func (s *ControllerClient) CreateOrUpdateDNSRecord(ctx context.Context, zone, name, rtype, content string, ttl int, proxy bool) error {
	record := DNSRequest{
		Zone:    zone,
		Name:    name,
		RType:   rtype,
		Content: content,
		TTL:     ttl,
		Proxy:   proxy,
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	req := &edgeproto.AccessDataRequest{
		Type: CreateOrUpdateDNSRecord,
		Data: data,
	}
	_, err = s.client.GetAccessData(ctx, req)
	return err
}

func (s *ControllerClient) GetDNSRecords(ctx context.Context, zone, fqdn string) ([]cloudflare.DNSRecord, error) {
	record := DNSRequest{
		Zone: zone,
		Name: fqdn,
	}
	data, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	req := &edgeproto.AccessDataRequest{
		Type: GetDNSRecords,
		Data: data,
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return nil, err
	}
	records := make([]cloudflare.DNSRecord, 0)
	err = json.Unmarshal(reply.Data, &records)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (s *ControllerClient) DeleteDNSRecord(ctx context.Context, zone, recordID string) error {
	record := DNSRequest{
		Zone: zone,
		Name: recordID,
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	req := &edgeproto.AccessDataRequest{
		Type: DeleteDNSRecord,
		Data: data,
	}
	_, err = s.client.GetAccessData(ctx, req)
	return err
}

func (s *ControllerClient) GetSessionTokens(ctx context.Context, arg []byte) (map[string]string, error) {
	req := &edgeproto.AccessDataRequest{
		Type: GetSessionTokens,
		Data: arg,
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return nil, err
	}
	tokens := map[string]string{}
	err = json.Unmarshal(reply.Data, &tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *ControllerClient) GetKafkaCreds(ctx context.Context) (*node.KafkaCreds, error) {
	req := &edgeproto.AccessDataRequest{
		Type: GetKafkaCreds,
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return nil, err
	}
	creds := node.KafkaCreds{}
	err = json.Unmarshal(reply.Data, &creds)
	return &creds, err
}

func (s *ControllerClient) GetGCSCreds(ctx context.Context) ([]byte, error) {
	req := &edgeproto.AccessDataRequest{
		Type: GetGCSCreds,
	}
	reply, err := s.client.GetAccessData(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply.Data, err
}
