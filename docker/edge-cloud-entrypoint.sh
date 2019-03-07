#!/bin/bash

set -e
source /root/mex-docker.env

if [ $# -le 0 ]; then
    echo "which program? argument required"
    exit 1
fi

case "$1" in
    controller)
	shift
	controller $*
	;;
    crmserver)
	shift
	crmserver $*
	;;
    dme-server)
	shift
	dme-server $*
	;;
    edgectl)
	shift
	edgectl $*
	;;
    loc-api-sim)
	shift
	loc-api-sim $*
	;;
    mc)
	shift
	mc $*
	;;
    tok-srv-sim)
	shift
	tok-srv-sim $*
	;;
    test-edgectl)
	shift
	test-edgectl.sh $*
	;;
    bash)
	shift
	/bin/bash $*
	;;
    *)
	echo invalid program $1
	exit 1
	;;
esac

