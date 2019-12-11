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

# enable approle auth if not already enabled
auths=$(vault auth list)
case "$auths" in
    *_"approle"_*) ;;
    *) vault auth enable approle
esac

# set up regional kv database
vault secrets enable -path=$REGION/jwtkeys kv
vault kv enable-versioning $REGION/jwtkeys
vault write $REGION/jwtkeys/config max_versions=2

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

path "secret/data/$REGION/cloudlet/openstack/*" {
  capabilities = [ "create", "update", "delete", "read" ]
}

path "secret/metadata/$REGION/cloudlet/openstack/*" {
  capabilities = [ "delete" ]
}

path "secret/data/cloudlet/openstack/mexenv.json" {
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
EOF
vault policy write $REGION.dme /tmp/dme-pol.hcl
rm /tmp/dme-pol.hcl
vault write auth/approle/role/$REGION.dme period="720h" policies="$REGION.dme"
# get dme app roleID and generate secretID
vault read auth/approle/role/$REGION.dme/role-id
vault write -f auth/approle/role/$REGION.dme/secret-id

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
