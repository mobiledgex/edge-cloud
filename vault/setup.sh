#!/bin/sh

# exit immediately on failure
set -e

# This collection of commands are used to configure Vault for e2e testing.
# The commands here represent a subset of configuration used for the
# production Vault. This script is not idempotent, so it cannot be re-run
# against the production Vault, but as new config is added, commands can
# copy-and-pasted to add new config to the production Vault.

# You may need to set the following env vars before running:
# export VAULT_ADDR=http://127.0.0.1:8200
# VAULT_TOKEN=<my auth token>

# enable approle auth
vault auth enable approle

# set up jwtkey key-value database
vault secrets enable -path=jwtkeys kv
vault kv enable-versioning jwtkeys
vault write jwtkeys/config max_versions=2

# these are commented out but are used to set the dme/mcorm secrets
#vault kv put jwtkeys/dme secret=12345 refresh=60m
#vault kv put jwtkeys/mcorm secret=12345 refresh=60m
#vault kv get jwtkeys/dme
#vault kv metadata get jwtkeys/dme

# set dme approle
cat > /tmp/dme-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "jwtkeys/data/dme" {
  capabilities = [ "read" ]
}

path "jwtkeys/metadata/dme" {
  capabilities = [ "read" ]
}
EOF
vault policy write dme /tmp/dme-pol.hcl
rm /tmp/dme-pol.hcl
vault write auth/approle/role/dme period="720h" policies="dme"
# get dme app roleID and generate secretID
vault read auth/approle/role/dme/role-id
vault write -f auth/approle/role/dme/secret-id

# set mcorm approle
cat > /tmp/mcorm-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "jwtkeys/data/mcorm" {
  capabilities = [ "read" ]
}

path "jwtkeys/metadata/mcorm" {
  capabilities = [ "read" ]
}
EOF
vault policy write mcorm /tmp/mcorm-pol.hcl
rm /tmp/mcorm-pol.hcl
vault write auth/approle/role/mcorm period="720h" policies="mcorm"
# get mcorm app roleID and generate secretID
vault read auth/approle/role/mcorm/role-id
vault write -f auth/approle/role/mcorm/secret-id

# set rotator approle - rotates dme/mcorm secret
cat > /tmp/rotator-pol.hcl <<EOF
path "auth/approle/login" {
  capabilities = [ "create", "read" ]
}

path "jwtkeys/data/*" {
  capabilities = [ "create", "update", "read" ]
}

path "jwtkeys/metadata/*" {
  capabilities = [ "read" ]
}
EOF
vault policy write rotator /tmp/rotator-pol.hcl
rm /tmp/rotator-pol.hcl
vault write auth/approle/role/rotator period="720h" policies="rotator"
# get rotator app roleID and generate secretID
vault read auth/approle/role/rotator/role-id
vault write -f auth/approle/role/rotator/secret-id

# generate secret string:
# openssl rand -base64 128

