// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloudcommon

import (
	"fmt"
	"strings"

	dme "github.com/edgexr/edge-cloud/d-match-engine/dme-proto"
	"github.com/edgexr/edge-cloud/edgeproto"
)

const (
	// Important: key strings used here for grpc metadata keys
	// MUST be lower-case.
	CallerAutoProv         = "caller-auto-prov"
	AutoProvReason         = "auto-prov-reason"
	AutoProvReasonDemand   = "demand"
	AutoProvReasonMinMax   = "minmax"
	AutoProvReasonOrphaned = "orphaned"
	AutoProvPolicyName     = "auto-prov-policy-name"
	AccessKeyData          = "access-key-data"
	AccessKeySig           = "access-key-sig"
	VaultKeySig            = "vault-key-sig"
)

var AutoProvMinAlreadyMetError = fmt.Errorf("Create to satisfy min already met, ignoring")

func AutoProvCloudletInfoOnline(cloudletInfo *edgeproto.CloudletInfo) bool {
	// Transitional states are considered "online".
	if cloudletInfo.State == dme.CloudletState_CLOUDLET_STATE_OFFLINE {
		return false
	}
	return true
}

func AutoProvCloudletOnline(cloudlet *edgeproto.Cloudlet) bool {
	// any maintenance state is considered offline
	if cloudlet.MaintenanceState != dme.MaintenanceState_NORMAL_OPERATION {
		return false
	}
	return true
}

func AutoProvAppInstOnline(appInst *edgeproto.AppInst, cloudletInfo *edgeproto.CloudletInfo, cloudlet *edgeproto.Cloudlet) bool {
	// Transitional states are considered "online"...but health check
	// doesn't actually have transitional states, except perhaps unknown.
	appInstOnline := false
	if appInst.HealthCheck == dme.HealthCheck_HEALTH_CHECK_UNKNOWN ||
		appInst.HealthCheck == dme.HealthCheck_HEALTH_CHECK_OK {
		appInstOnline = true
	}
	return appInstOnline && AutoProvCloudletInfoOnline(cloudletInfo) && AutoProvCloudletOnline(cloudlet)
}

func AutoProvAppInstGoingOnline(appInst *edgeproto.AppInst, cloudletInfo *edgeproto.CloudletInfo, cloudlet *edgeproto.Cloudlet) bool {
	appInstGoingOnline := false
	if appInst.State == edgeproto.TrackedState_CREATE_REQUESTED || appInst.State == edgeproto.TrackedState_CREATING || appInst.State == edgeproto.TrackedState_CREATING_DEPENDENCIES {
		appInstGoingOnline = true
	}
	return appInstGoingOnline && AutoProvCloudletInfoOnline(cloudletInfo) && AutoProvCloudletOnline(cloudlet)
}

func AppInstBeingDeleted(inst *edgeproto.AppInst) bool {
	if inst.State == edgeproto.TrackedState_DELETE_REQUESTED || inst.State == edgeproto.TrackedState_DELETING || inst.State == edgeproto.TrackedState_DELETE_DONE || inst.State == edgeproto.TrackedState_NOT_PRESENT {
		return true
	}
	return false
}

const (
	AlreadyUnderDeletionMsg          = "busy, already under deletion"
	StreamActionAlreadyInProgressMsg = "An action is already in progress for the object"
)

// Autoprov relies on detecting if an AppInst is already being created
func IsAppInstBeingCreatedError(err error) bool {
	if strings.Contains(err.Error(), "AppInst key") && strings.Contains(err.Error(), "already exists") {
		// obj.ExistsError()
		return true
	}
	if strings.Contains(err.Error(), StreamActionAlreadyInProgressMsg) {
		// stream autocluster error
		return true
	}
	return false
}

// Autoprov relies on detecting if an AppInst is already being deleted
func IsAppInstBeingDeletedError(err error) bool {
	if strings.Contains(err.Error(), AlreadyUnderDeletionMsg) {
		return true
	}
	if strings.Contains(err.Error(), StreamActionAlreadyInProgressMsg) {
		// stream autocluster error
		return true
	}
	return false
}

// Cluster name to trigger using an existing free reservable ClusterInst
// or creating a new one automatically.
// Because this name is always part of the AppInstKey in etcd,
// and because AutoProv will only ever instantiate once instance
// of an App per cloudlet, there are really no uniqueness requirements
// on this name.
// Additionally any objects instantiated at the infra level that
// are independent of the AppInst key should be using the
// real cluster name from the ClusterInst object.
var AutoProvClusterName = AutoClusterPrefix + "-autoprov"
