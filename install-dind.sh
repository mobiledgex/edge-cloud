#!/bin/bash

if [ ! -e  /usr/local/bin/dind-cluster-v1.14.sh ]; then
   wget https://github.com/kubernetes-sigs/kubeadm-dind-cluster/releases/download/v0.1.0/dind-cluster-v1.14.sh
   mv dind-cluster-v1.14.sh /usr/local/bin/
   chmod +x /usr/local/bin/dind-cluster-v1.14.sh
else
   echo "dind-cluster-v1.14.sh already installed"
fi
