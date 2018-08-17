#!/bin/bash
edgectl controller CreateFlavor --key-name x1.medium --ram 8000000 --vcpus 4
edgectl controller CreateDeveloper --key-name testdeveloper --address '111 ave' --email dev@g.com --key-name testdeveloper --passhash 999 --username testdeveloper
edgectl controller CreateOperator --key-name mex-gddt
edgectl controller CreateCloudlet --key-name testcloudlet --key-operatorkey-name mex-gddt --location-altitude 1.1 --location-long 1.1 --location-lat 1.1  --numdynamicips 1
edgectl controller CreateCluster --defaultflavor-name x1.medium --key-name testcluster 
edgectl controller CreateClusterFlavor --key-name x1.medium --masterflavor-name x1.medium --maxnodes 2 --nodeflavor-name x1.medium --nummasters 1 --numnodes 2
edgectl controller CreateClusterInst --key-cloudletkey-operatorkey-name mex-gddt --key-cloudletkey-name testcloudlet --key-clusterkey-name testcluster 
edgectl controller CreateApp --accesslayer AccessLayerL7 --accessports tcp:27272,tcp:27273,tcp:27274 --cluster-name testcluster --config http://registry.mobiledgex.net:8080/mobiledgex/testapp.yaml --defaultflavor-name x1.medium --imagetype ImageTypeDocker  --key-developerkey-name  testdeveloper --key-name testapp --key-version testversion
edgectl controller CreateAppInst --accesslayer AccessLayerL7 --key-appkey-developerkey-name testdeveloper --key-appkey-name testapp --key-appkey-version testversion --key-cloudletkey-operatorkey-name mex-gddt --key-cloudletkey-name testcloudlet --key-id 1
