#!/bin/sh

# exit immediately on failure
set -e

# Set up the profiles for the edge-cloud approles.
# This assumes a global Vault for all regions, so paths in the Vault
# are region-specific.
# This script should be run for each new region that we bring online.

# You may need to set the following env vars before running:
# VAULT_ADDR=http://127.0.0.1:8200
# VAULT_TOKEN=<my auth token>

# Region should be set to the correct region name
# REGION=local
REGION=$1

if [ -z "$REGION" ]; then
    echo "Usage: setup-region.sh <region>"
    exit 1
fi
echo "Setting up Vault region $REGION"

# set up regional kv database
vault secrets enable -path=$REGION/jwtkeys kv
vault kv enable-versioning $REGION/jwtkeys
vault write $REGION/jwtkeys/config max_versions=2

# set up regional cert issuer role
vault write pki-regional/roles/$REGION \
      allow_localhost=true \
      allowed_domains="mobiledgex.net" \
      allow_subdomains=true \
      allowed_uri_sans="region://$REGION"

# set up cloudlet regional cert issuer role
vault write pki-regional-cloudlet/roles/$REGION \
      allow_localhost=true \
      allowed_domains="mobiledgex.net" \
      allow_subdomains=true \
      allowed_uri_sans="region://$REGION"

cat > /tmp/controller-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "secret/data/registry/*" {
  capabilities = [ "read" ]
}

path "secret/data/$REGION/cloudlet/*" {
  capabilities = [ "create", "update", "delete", "read" ]
}

path "secret/data/cloudlet/*" {
  capabilities = [ "read" ]
}

path "secret/data/$REGION/accounts/*" {
  capabilities = [ "read" ]
}

path "pki-regional/issue/$REGION" {
  capabilities = [ "read", "update" ]
}
EOF
vault policy write $REGION.controller /tmp/controller-pol.hcl
rm /tmp/controller-pol.hcl
vault write auth/approle/role/$REGION.controller period="720h" policies="$REGION.controller"
# get controller app roleID and generate secretID
vault read auth/approle/role/$REGION.controller/role-id
vault write -f auth/approle/role/$REGION.controller/secret-id

# set crm approle
cat > /tmp/crm-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "secret/data/registry/*" {
  capabilities = [ "read" ]
}

path "secret/data/accounts/chef" {
  capabilities = [ "read" ]
}

path "secret/data/$REGION/cloudlet/openstack/*" {
  capabilities = [ "create", "update", "delete", "read" ]
}

path "secret/metadata/$REGION/cloudlet/openstack/*" {
  capabilities = [ "delete" ]
}

path "secret/data/cloudlet/openstack/mexenv.json" {
  capabilities = [ "read" ]
}

path "pki-regional-cloudlet/issue/$REGION" {
  capabilities = [ "read", "update" ]
}

path "secret/data/keys/id_rsa_mex" {
  capabilities = [ "read" ]
}
EOF
vault policy write $REGION.crm /tmp/crm-pol.hcl
rm /tmp/crm-pol.hcl
vault write auth/approle/role/$REGION.crm period="720h" policies="$REGION.crm"
# get crm app roleID and generate secretID
vault read auth/approle/role/$REGION.crm/role-id
vault write -f auth/approle/role/$REGION.crm/secret-id

# set dme approle
cat > /tmp/dme-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "$REGION/jwtkeys/data/dme" {
  capabilities = [ "read" ]
}

path "$REGION/jwtkeys/metadata/dme" {
  capabilities = [ "read" ]
}

# Allow access to certs (including access to cert creation)
path "certs/*" {
  capabilities = ["read"]
}

path "pki-regional-cloudlet/issue/$REGION" {
  capabilities = [ "read", "update" ]
}
EOF
vault policy write $REGION.dme /tmp/dme-pol.hcl
rm /tmp/dme-pol.hcl
vault write auth/approle/role/$REGION.dme period="720h" policies="$REGION.dme"
# get dme app roleID and generate secretID
vault read auth/approle/role/$REGION.dme/role-id
vault write -f auth/approle/role/$REGION.dme/secret-id

# set cluster-svc approle
cat > /tmp/cluster-svc-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "pki-regional/issue/$REGION" {
  capabilities = [ "read", "update" ]
}
EOF
vault policy write $REGION.cluster-svc /tmp/cluster-svc-pol.hcl
rm /tmp/cluster-svc-pol.hcl
vault write auth/approle/role/$REGION.cluster-svc period="720h" policies="$REGION.cluster-svc"
# get cluster-svc app roleID and generate secretID
vault read auth/approle/role/$REGION.cluster-svc/role-id
vault write -f auth/approle/role/$REGION.cluster-svc/secret-id

# set rotator approle - rotates dme secret
cat > /tmp/rotator-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "$REGION/jwtkeys/data/*" {
  capabilities = [ "create", "update", "read" ]
}

path "$REGION/jwtkeys/metadata/*" {
  capabilities = [ "read" ]
}
EOF
vault policy write $REGION.rotator /tmp/rotator-pol.hcl
rm /tmp/rotator-pol.hcl
vault write auth/approle/role/$REGION.rotator period="720h" policies="$REGION.rotator"
# get rotator app roleID and generate secretID
vault read auth/approle/role/$REGION.rotator/role-id
vault write -f auth/approle/role/$REGION.rotator/secret-id

# generate secret string:
# openssl rand -base64 128

# Generate regional cert for edgectl
mkdir -p /tmp/edgectl.$REGION
vault write -format=json pki-regional/issue/$REGION \
      common_name=edgectl.mobiledgex.net \
      alt_names=localhost \
      ip_sans="127.0.0.1,0.0.0.0" \
      uri_sans="region://$REGION" > /tmp/edgectl.$REGION/issue
cat /tmp/edgectl.$REGION/issue | jq -r .data.certificate > /tmp/edgectl.$REGION/mex.crt
cat /tmp/edgectl.$REGION/issue | jq -r .data.private_key > /tmp/edgectl.$REGION/mex.key
cat /tmp/edgectl.$REGION/issue | jq -r .data.issuing_ca > /tmp/edgectl.$REGION/mex-ca.crt

# set edgeturn approle
cat > /tmp/edgeturn-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "pki-regional/issue/$REGION" {
  capabilities = [ "read", "update" ]
}
EOF
vault policy write $REGION.edgeturn /tmp/edgeturn-pol.hcl
rm /tmp/edgeturn-pol.hcl
vault write auth/approle/role/$REGION.edgeturn period="720h" policies="$REGION.edgeturn"
# get edgeturn app roleID and generate secretID
vault read auth/approle/role/$REGION.edgeturn/role-id
vault write -f auth/approle/role/$REGION.edgeturn/secret-id
