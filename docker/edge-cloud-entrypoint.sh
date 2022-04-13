#!/bin/bash
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


set -e
source /root/mex-docker.env

if [ $# -le 0 ]; then
    echo "which program? argument required"
    exit 1
fi

case "$1" in
    controller|\
    cluster-svc|\
    crmserver|\
    dme-server|\
    edgectl|\
    edgeturn|\
    frm|\
    loc-api-sim|\
    mc|\
    mcctl|\
    notifyroot|\
    shepherd|\
    alertmgr-sidecar|\
    tok-srv-sim)
	"$@"
	;;
    test-edgectl)
	shift
	test-edgectl.sh "$@"
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
	/bin/bash "$@"
	;;
    *)
	echo invalid program $1
	exit 1
	;;
esac

