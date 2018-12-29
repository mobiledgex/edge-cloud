#!/bin/bash

#set -e
if [ $# -le 0 ]; then
    echo 'create or remove?'
    exit 1
fi


function createall {
    edgectl controller CreateFlavor --key-name x1.medium --ram 8000000 --vcpus 4 --disk 1
    edgectl controller CreateDeveloper --key-name testdeveloper --address '111 ave' --email dev@g.com --key-name testdeveloper --passhash 999 --username testdeveloper
    edgectl controller CreateOperator --key-name gddt
    edgectl controller CreateCloudlet --key-name sunnydale-test --key-operatorkey-name gddt --location-altitude 1.1 --location-long 1.1 --location-lat 1.1  --numdynamicips 1
    edgectl controller CreateClusterFlavor --key-name x1.medium --masterflavor-name x1.medium --maxnodes 2 --nodeflavor-name x1.medium --nummasters 1 --numnodes 2
    edgectl controller CreateCluster --defaultflavor-name x1.medium --key-name sunnydale-test
    edgectl controller CreateClusterInst --key-cloudletkey-operatorkey-name gddt --key-cloudletkey-name sunnydale-test --key-clusterkey-name sunnydale-test 
    edgectl controller CreateApp  --accessports tcp:27272,tcp:27273,tcp:27274,tcp:27275,udp:27276 --cluster-name sunnydale-test  --defaultflavor-name x1.medium --imagetype ImageTypeDocker  --key-developerkey-name  testdeveloper --key-name mytest-app --key-version testversion --deploymentmanifest kustomize/application/output/mytest-app.yaml
    edgectl controller CreateAppInst  --key-appkey-developerkey-name testdeveloper --key-appkey-name mytest-app --key-appkey-version testversion --key-cloudletkey-operatorkey-name gddt --key-cloudletkey-name sunnydale-test --key-id 1
}

function removeall {
    edgectl controller DeleteAppInst  --key-appkey-developerkey-name testdeveloper --key-appkey-name mytest-app --key-appkey-version testversion --key-cloudletkey-operatorkey-name gddt --key-cloudletkey-name sunnydale-test --key-id 1
    edgectl controller DeleteApp  --accessports tcp:27272,tcp:27273,tcp:27274,tcp:27275,udp:27276 --cluster-name sunnydale-test  --defaultflavor-name x1.medium --imagetype ImageTypeDocker  --key-developerkey-name  testdeveloper --key-name mytest-app --key-version testversion --deploymentmanifest kustomize/application/output/mytest-app.yaml
    edgectl controller DeleteClusterFlavor --key-name x1.medium --masterflavor-name x1.medium --maxnodes 2 --nodeflavor-name x1.medium --nummasters 1 --numnodes 2
    edgectl controller DeleteClusterInst --key-cloudletkey-operatorkey-name gddt --key-cloudletkey-name sunnydale-test --key-clusterkey-name sunnydale-test
    edgectl controller DeleteCluster --defaultflavor-name x1.medium --key-name sunnydale-test
    edgectl controller DeleteCloudlet --key-name sunnydale-test --key-operatorkey-name gddt --location-altitude 1.1 --location-long 1.1 --location-lat 1.1  --numdynamicips 1
    edgectl controller DeleteOperator --key-name gddt
    edgectl controller DeleteDeveloper --key-name testdeveloper --address '111 ave' --email dev@g.com --key-name testdeveloper --passhash 999 --username testdeveloper
    edgectl controller DeleteFlavor --key-name x1.medium --ram 8000000 --vcpus 4 --disk 1
}


case "$1" in
    create)
	shift
	createall
	;;
    remove)
	shift
	removeall
	;;
    *)
	echo invalid command, need create or remove
	exit 1
	;;
esac
