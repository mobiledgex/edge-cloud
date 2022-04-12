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


#default to use env vars
secret1=$DMESECRET
secret2=$MCORMSECRET

# need 2 args, for DME an
if [ $# -eq 2 ]; then
   secret1=$1
   secret2=$2 
fi

if [ -z "$secret1" ]; then
   echo "dme secret not set by parm or env var"
   exit 1
fi

if [ -z "$secret2" ]; then
   echo "mcorm  secret not set by parm or env var"
   exit 1
fi


echo "running Mex vault setup"
/root/setup.sh

echo "Setting secrets"
## first put old secret and rotate it
vault kv put jwtkeys/dme secret=$secret1-old refresh=60m
vault kv put jwtkeys/mcorm secret=$secret2-old refresh=60m

#put the current secret
vault kv put jwtkeys/dme secret=$secret1 refresh=60m
vault kv put jwtkeys/mcorm secret=$secret2 refresh=60m

dmerole=`vault read auth/approle/role/dme/role-id|grep role_id|tr -s " " |cut -d " " -f 2`
dmesecret=`vault write -f auth/approle/role/dme/secret-id|grep "secret_id " |tr -s " " |cut -d " " -f 2`

## for use cut and paste into k8s manifest
echo "k8s manifest values for DME:"
echo ""
echo "         - name: VAULT_ROLE_ID"
echo "           value: $dmerole"
echo "         - name: VAULT_SECRET_ID"
echo "           value: $dmesecret"
