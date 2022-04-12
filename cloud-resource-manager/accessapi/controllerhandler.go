// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	case GetPublicCert:
		publicCert, err := s.vaultClient.GetPublicCert(ctx, string(req.Data))
		if err != nil {
			return nil, err
		}
		out, merr = json.Marshal(*publicCert)
	case GetKafkaCreds:
		creds, err := s.vaultClient.GetKafkaCreds(ctx)
		if err != nil {
			return nil, err
		}
		out, merr = json.Marshal(creds)
	case GetGCSCreds:
		creds, err := s.vaultClient.GetGCSCreds(ctx)
		if err != nil {
			return nil, err
		}
		out = creds
	case GetFederationAPIKey:
		apiKey, err := s.vaultClient.GetFederationAPIKey(ctx, string(req.Data))
		if err != nil {
			return nil, err
		}
		out = []byte(apiKey)
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
