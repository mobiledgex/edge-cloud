package accessapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Handles unmarshaling of data from ControllerClient. It then calls
// to the VaultClient to access data from Vault.
type ControllerHandler struct {
	vaultClient *VaultClient
}

func NewControllerHandler(vaultClient *VaultClient) *ControllerHandler {
	return &ControllerHandler{
		vaultClient: vaultClient,
	}
}

func (s *ControllerHandler) GetAccessData(ctx context.Context, req *edgeproto.AccessDataRequest) (*edgeproto.AccessDataReply, error) {
	var out []byte
	var merr error
	switch req.Type {
	case GetCloudletAccessVars:
		vars, err := s.vaultClient.GetCloudletAccessVars(ctx)
		if err != nil {
			return nil, err
		}
		out, merr = json.Marshal(vars)
	case GetRegistryAuth:
		auth, err := s.vaultClient.GetRegistryAuth(ctx, string(req.Data))
		if err != nil {
			return nil, err
		}
		out, merr = json.Marshal(auth)
	case SignSSHKey:
		signed, err := s.vaultClient.SignSSHKey(ctx, string(req.Data))
		if err != nil {
			return nil, err
		}
		out = []byte(signed)
	case GetSSHPublicKey:
		pubkey, err := s.vaultClient.GetSSHPublicKey(ctx)
		if err != nil {
			return nil, err
		}
		out = []byte(pubkey)
	case GetOldSSHKey:
		mexkey, err := s.vaultClient.GetOldSSHKey(ctx)
		if err != nil {
			return nil, err
		}
		out, merr = json.Marshal(mexkey)
	case GetChefAuthKey:
		auth, err := s.vaultClient.GetChefAuthKey(ctx)
		if err != nil {
			return nil, err
		}
		out, merr = json.Marshal(auth)
	case CreateOrUpdateDNSRecord:
		dnsReq := DNSRequest{}
		err := json.Unmarshal(req.Data, &dnsReq)
		if err != nil {
			return nil, err
		}
		err = s.vaultClient.CreateOrUpdateDNSRecord(ctx, dnsReq.Zone, dnsReq.Name, dnsReq.RType, dnsReq.Content, dnsReq.TTL, dnsReq.Proxy)
		if err != nil {
			return nil, err
		}
	case GetDNSRecords:
		dnsReq := DNSRequest{}
		err := json.Unmarshal(req.Data, &dnsReq)
		if err != nil {
			return nil, err
		}
		records, err := s.vaultClient.GetDNSRecords(ctx, dnsReq.Zone, dnsReq.Name)
		if err != nil {
			return nil, err
		}
		out, merr = json.Marshal(records)
	case DeleteDNSRecord:
		dnsReq := DNSRequest{}
		err := json.Unmarshal(req.Data, &dnsReq)
		if err != nil {
			return nil, err
		}
		err = s.vaultClient.DeleteDNSRecord(ctx, dnsReq.Zone, dnsReq.Name)
		if err != nil {
			return nil, err
		}
	case GetSessionTokens:
		tokens, err := s.vaultClient.GetSessionTokens(ctx, req.Data)
		if err != nil {
			return nil, err
		}
		out, merr = json.Marshal(tokens)
	default:
		return nil, fmt.Errorf("Unexpected request data type %s", req.Type)
	}
	if merr != nil {
		return nil, merr
	}
	return &edgeproto.AccessDataReply{
		Data: out,
	}, nil
}
