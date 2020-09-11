#!/bin/sh

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

CADIR=/tmp/vault_pki

rm -Rf $CADIR
mkdir $CADIR

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
