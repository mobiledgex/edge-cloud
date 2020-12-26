#!/bin/bash

# exit immediately on failure
set -e

USAGE="usage: $( basename $0 ) <options>

 -d <certs-dir>         TLS Certs directory
 -c <cert-file>         TLS Cert file
 -k <key-file>          TLS Key file

 -h                    Display this help message
"

while getopts ":hd:c:k:" OPT; do
        case "$OPT" in
        h) echo "$USAGE"; exit 0 ;;
        d) CERTS_DIR="$OPTARG" ;;
        c) CERT_FILE="$OPTARG" ;;
        k) KEY_FILE="$OPTARG" ;;
        esac
done
shift $(( OPTIND - 1 ))

die() {
        echo "ERROR: $*" >&2
        exit 2
}

[[ -z $CERTS_DIR ]] && die "Missing argument '-d'"
[[ -z $CERT_FILE ]] && die "Missing argument '-c'"
[[ -z $KEY_FILE ]] && die "Missing argument '-k'"

cd $CERTS_DIR

# Old format, move to using double symlinks
# also upgrade old containers to use new envoy image
if [[ -f $CERT_FILE ]]; then
        OLD_CERTS_DIR=..certs_$(date "+%Y_%m_%d_%H_%M_%S.%s")
        mkdir -p $OLD_CERTS_DIR
        mv $CERT_FILE $OLD_CERTS_DIR/
        mv $KEY_FILE $OLD_CERTS_DIR/
        ln -sf $OLD_CERTS_DIR ..data
        ln -sf ..data/$CERT_FILE $CERT_FILE
        ln -sf ..data/$KEY_FILE $KEY_FILE

	for envoyName in $(docker ps --format "{{.Names}}" --filter name="^envoy"); do
	  envoyPath="/home/ubuntu/envoy/${envoyName#envoy}"
	  if ! grep -iq "tls_context" $envoyPath/envoy.yaml; then
		  # skip if envoy.yaml is not configured with TLS
		  continue
	  fi

	  # patch envoy.yaml to use sds
	  sed -i "1 s/^/node:\n  id: ${envoyName}\n  cluster: ${envoyName}/" $envoyPath/envoy.yaml

	  sed -i '/      tls_context:/,/.key"/c\
      transport_socket:\
        name: "envoy.transport_sockets.tls"\
        typed_config:\
          "@type": "type.googleapis.com/envoy.api.v2.auth.DownstreamTlsContext"\
          common_tls_context:\
            tls_certificate_sds_secret_configs:\
                sds_config:\
                    path: '"${envoyPath}"'/sds.yaml' $envoyPath/envoy.yaml

	  # write sds.yaml
	  cat > $envoyPath/sds.yaml <<EOF
resources:
- "@type": "type.googleapis.com/envoy.api.v2.auth.Secret"
  tls_certificates:
    certificate_chain:
      filename: "/etc/envoy/certs/envoyTlsCerts.crt"
    private_key:
      filename: "/etc/envoy/certs/envoyTlsCerts.key"
EOF

	  # stop and start docker with new image
	  docker stop $envoyName
	  docker rm $envoyName
	  runcmd=$(docker run --rm -v /var/run/docker.sock:/var/run/docker.sock assaflavie/runlike $envoyName)
	  runcmd_newimg=$(sed 's/envoy-with-curl.*\? /envoy-with-curl@sha256:9bc06553ad6add6bfef1d8a1b04f09721415975e2507da0a2d5b914c066474df /g' <<< $runcmd)
	  echo "$envoyName=>$out"
  	  $($runcmd_newimg)
	done
fi

# Atomic update of certs using symlinks
if [[ -f $CERT_FILE.new ]]; then
        NEW_CERTS_DIR=..certs_$(date "+%Y_%m_%d_%H_%M_%S.%s")
        mkdir -p $NEW_CERTS_DIR
        mv $CERT_FILE.new $NEW_CERTS_DIR/$CERT_FILE
        mv $KEY_FILE.new $NEW_CERTS_DIR/$KEY_FILE
        rm -f ..tmp; ln -sf $NEW_CERTS_DIR ..tmp
        mv -Tf ..tmp ..data

        # Prune old certs
        find -type d -name "..certs_*" -not -path "./$NEW_CERTS_DIR" | xargs rm -rf
fi
