#!/bin/sh
# Copyright 2022 MobiledgeX, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# exit immediately on failure
set -e

# Set up the global settings for Vault.

# You may need to set the following env vars before running:
# VAULT_ADDR=http://127.0.0.1:8200
# VAULT_TOKEN=<my auth token>

echo "Setting up Vault"

# enable approle if not already enabled
auths=$(vault auth list)
case "$auths" in
    *_"approle"_*) ;;
    *) vault auth enable approle
esac

if [ -z $CADIR ]; then
    CADIR=/tmp/vault_pki
fi

rm -Rf $CADIR
mkdir -p $CADIR

# enable root pki
vault secrets enable pki
# generate root cert
vault write -format=json pki/root/generate/internal \
      common_name=localhost | jq -r '.data.certificate' > $CADIR/rootca.pem
vault write pki/config/urls \
    issuing_certificates="$VAULT_ADDR/v1/pki/ca" \
    crl_distribution_points="$VAULT_ADDR/v1/pki/crl"

# enable global intermediate pki
vault secrets enable -path=pki-global pki
vault secrets tune -max-lease-ttl=72h pki-global
vault write pki-global/config/urls \
    issuing_certificates="$VAULT_ADDR/v1/pki-global/ca" \
    crl_distribution_points="$VAULT_ADDR/v1/pki-global/crl"
# generate intermediate cert
vault write -format=json pki-global/intermediate/generate/internal \
      common_name="pki-global" | jq -r '.data.csr' > $CADIR/pki_global.csr
# sign intermediate with root
vault write -format=json pki/root/sign-intermediate csr=@$CADIR/pki_global.csr \
      format=pem_bundle | jq -r '.data.certificate' > $CADIR/global.cert.pem
# imported signed intermediate cert
vault write pki-global/intermediate/set-signed certificate=@$CADIR/global.cert.pem

# enable regional secure intermediate pki
vault secrets enable -path=pki-regional pki
vault secrets tune -max-lease-ttl=72h pki-regional
vault write pki-regional/config/urls \
    issuing_certificates="$VAULT_ADDR/v1/pki-regional/ca" \
    crl_distribution_points="$VAULT_ADDR/v1/pki-regional/crl"
# generate intermediate cert
vault write -format=json pki-regional/intermediate/generate/internal \
      common_name="pki-regional" | jq -r '.data.csr' > $CADIR/pki_regional.csr
# sign intermediate with root
vault write -format=json pki/root/sign-intermediate csr=@$CADIR/pki_regional.csr \
      format=pem_bundle | jq -r '.data.certificate' > $CADIR/regional.cert.pem
# imported signed intermediate cert
vault write pki-regional/intermediate/set-signed certificate=@$CADIR/regional.cert.pem

# enable regional cloudlet intermediate pki
vault secrets enable -path=pki-regional-cloudlet pki
vault secrets tune -max-lease-ttl=72h pki-regional-cloudlet
vault write pki-regional-cloudlet/config/urls \
    issuing_certificates="$VAULT_ADDR/v1/pki-regional-cloudlet/ca" \
    crl_distribution_points="$VAULT_ADDR/v1/pki-regional-cloudlet/crl"
# generate intermediate cert
vault write -format=json pki-regional-cloudlet/intermediate/generate/internal \
      common_name="pki-regional-cloudlet" | jq -r '.data.csr' > $CADIR/pki_cloudlet_regional.csr
# sign intermediate with root
vault write -format=json pki/root/sign-intermediate csr=@$CADIR/pki_cloudlet_regional.csr \
      format=pem_bundle | jq -r '.data.certificate' > $CADIR/cloudlet.regional.cert.pem
# imported signed intermediate cert
vault write pki-regional-cloudlet/intermediate/set-signed certificate=@$CADIR/cloudlet.regional.cert.pem

# set up global cert issuer role
vault write pki-global/roles/default \
      allow_localhost=true \
      allowed_domains="mobiledgex.net" \
      allow_subdomains=true \
      allowed_uri_sans="region://none"

# set notifyroot approle - note this is a global service
cat > /tmp/notifyroot-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "pki-global/issue/*" {
  capabilities = [ "read", "update" ]
}
EOF
vault policy write notifyroot /tmp/notifyroot-pol.hcl
rm /tmp/notifyroot-pol.hcl
vault write auth/approle/role/notifyroot period="720h" policies="notifyroot"
# get notifyroot app roleID and generate secretID
vault read auth/approle/role/notifyroot/role-id
vault write -f auth/approle/role/notifyroot/secret-id

# enable vault ssh secrets engine
vault secrets enable -path=ssh ssh
vault write ssh/config/ca generate_signing_key=true
vault write ssh/roles/machine -<<"EOH"
{
  "allow_user_certificates": true,
  "allowed_users": "*",
  "allowed_extensions": "permit-pty,permit-port-forwarding",
  "default_extensions": [
    {
      "permit-pty": "",
      "permit-port-forwarding": ""
    }
  ],
  "key_type": "ca",
  "default_user": "ubuntu",
  "ttl": "72h",
  "max_ttl": "72h"
}
EOH
vault write ssh/roles/user -<<"EOH"
{
  "allow_user_certificates": true,
  "allowed_users": "*",
  "allowed_extensions": "permit-pty,permit-port-forwarding",
  "default_extensions": [
    {
      "permit-pty": "",
      "permit-port-forwarding": ""
    }
  ],
  "key_type": "ca",
  "default_user": "ubuntu",
  "ttl": "5m",
  "max_ttl": "60m"
}
EOH
