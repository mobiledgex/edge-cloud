#!/bin/sh

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
vault kv put jwtkeys/dme secret=$secret1 refresh=60m
vault kv put jwtkeys/mcorm secret=$secret2 refresh=60m

dmerole=`vault read auth/approle/role/dme/role-id|grep role_id|tr -s " " |cut -d " " -f 2`
echo DMEROLEID=$dmerole
dmesecret=`vault write -f auth/approle/role/dme/secret-id|grep "secret_id " |tr -s " " |cut -d " " -f 2`
echo DMESECRETID=$dmesecret
