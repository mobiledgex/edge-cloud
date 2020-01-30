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
    cluster-svc)
	shift
	cluster-svc $*
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
    mcctl)
	shift
	mcctl $*
	;;
    tok-srv-sim)
	shift
	tok-srv-sim $*
	;;
    test-edgectl)
	shift
	test-edgectl.sh $*
	;;
    shepherd)
	shift
	shepherd $*
	;;
    dump-docs)
	shift
	case "$1" in
	    client)	cat /usr/local/doc/client/app-client.swagger.json ;;
	    internal)	cat /usr/local/doc/internal/apidocs.swagger.json ;;
	    mc)		cat /usr/local/doc/mc/apidocs.swagger.json ;;
	    *)		cat /usr/local/doc/external/apidocs.swagger.json ;;
	esac
	;;
    version)
	shift
	cat /version.txt
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

