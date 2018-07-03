#!/bin/sh

set -e

if [ $# -le 0 ]; then
	echo "which program? argument required"
	exit 1
fi

case "$1" in
	controller)
		shift
		controller $*
		;;
	crmctl)
		shift
		crmctl $*
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
	*)
		echo invalid program $1
		exit 1
		;;
esac

