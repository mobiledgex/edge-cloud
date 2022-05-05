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

package main

import (
	"fmt"
	"strconv"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/edgexr/edge-cloud/cloudcommon"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/util"
)

// Functions to manage unique DNS labels.
// Note that AppInst and ClusterInst share the same namespace,
// as they are both used to compute the fully qualified domain name of
// format cloudletobject.cloudlet.region.domain.

// reserved cloudlet object names
var reservedCloudletObjectDnsLabels = map[string]struct{}{
	cloudcommon.RootLBHostname: struct{}{},
}

func (s *CloudletApi) setDnsLabel(stm concurrency.STM, cloudlet *edgeproto.Cloudlet) error {
	// More likely unique names should come first
	// to avoid being truncated.
	// To avoid the org from being completely truncated
	// if the cloudlet name is very long, limit the length
	// of the cloudlet name.
	name := dnsSanitizeTrunc(cloudlet.Key.Name, 40)
	oper := dnsSanitizeTrunc(cloudlet.Key.Organization, 20)
	baseLabel := name + "-" + oper

	// Number of iterations must be fairly low to avoid STM limits
	cloudlet.DnsLabel = ""
	for ii := 0; ii < 10; ii++ {
		label := genNextDnsLabel(baseLabel, cloudcommon.DnsCloudletLabelMaxLen, ii)
		if isReservedCloudletObjectDnsLabel(label) || s.dnsLabelStore.STMHas(stm, label) {
			continue
		}
		cloudlet.DnsLabel = label
		return nil
	}
	return dnsLabelError(baseLabel)
}

func (s *ClusterInstApi) setDnsLabel(stm concurrency.STM, ci *edgeproto.ClusterInst) error {
	// More likely unique names should come first
	// to avoid being truncated.
	name := dnsSanitizeTrunc(ci.Key.ClusterKey.Name, 40)
	org := dnsSanitizeTrunc(ci.Key.Organization, 20)
	baseLabel := name + "-" + org

	// Number of iterations must be fairly low to avoid STM limits
	ci.DnsLabel = ""
	for ii := 0; ii < 10; ii++ {
		label := genNextDnsLabel(baseLabel, cloudcommon.DnsCloudletObjectLabelMaxLen, ii)
		if isReservedCloudletObjectDnsLabel(label) || s.dnsLabelStore.STMHas(stm, &ci.Key.CloudletKey, label) {
			continue
		}
		ci.DnsLabel = label
		return nil
	}
	return dnsLabelError(baseLabel)
}

func (s *AppInstApi) setDnsLabel(stm concurrency.STM, ai *edgeproto.AppInst) error {
	// More likely unique names should come first
	// to avoid being truncated. Truncate fields separately
	// to avoid the last field from being completely truncated
	// if other fields are too long.
	app := dnsSanitizeTrunc(ai.Key.AppKey.Name, 60)
	ver := dnsSanitizeTrunc(ai.Key.AppKey.Version, 60)
	org := dnsSanitizeTrunc(ai.Key.AppKey.Organization, 60)
	baseLabel := app + ver + "-" + org

	if len(baseLabel) > cloudcommon.DnsCloudletObjectLabelMaxLen {
		// prioritize truncation of appname+version
		app = dnsSanitizeTrunc(app, 40)
		ver = dnsSanitizeTrunc(ver, 10)
		baseLabel = app + ver + "-" + org
	}

	// Number of iterations must be fairly low to avoid STM limits
	ai.DnsLabel = ""
	for ii := 0; ii < 10; ii++ {
		label := genNextDnsLabel(baseLabel, cloudcommon.DnsCloudletObjectLabelMaxLen, ii)
		if isReservedCloudletObjectDnsLabel(label) || s.dnsLabelStore.STMHas(stm, &ai.Key.ClusterInstKey.CloudletKey, label) {
			continue
		}
		ai.DnsLabel = label
		return nil
	}
	return dnsLabelError(baseLabel)
}

func dnsSanitizeTrunc(name string, maxLen int) string {
	name = util.DNSSanitize(name)
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen]
}

func isReservedCloudletObjectDnsLabel(label string) bool {
	_, found := reservedCloudletObjectDnsLabels[label]
	return found
}

func genNextDnsLabel(label string, maxLen, counter int) string {
	suffix := ""
	if counter > 0 {
		suffix = strconv.Itoa(counter)
	}
	if len(suffix) >= maxLen {
		panic(fmt.Sprintf("suffix %s cannot be longer than dns segment max len of %d", suffix, maxLen))
	}
	truncBy := len(label) + len(suffix) - maxLen
	if truncBy > 0 {
		label = label[:len(label)-truncBy]
	}
	return label + suffix
}

func dnsLabelError(baseLabel string) error {
	return fmt.Errorf("Unable to compute unique DNS label from base label of %q, please change key values", baseLabel)
}

func getCloudletRootLBFQDN(cloudlet *edgeproto.Cloudlet) string {
	reg := util.HostnameSanitize(*region)
	return fmt.Sprintf("%s.%s.%s.%s", cloudcommon.RootLBHostname, cloudlet.DnsLabel, reg, *appDNSRoot)
}

func getClusterInstFQDN(ci *edgeproto.ClusterInst, cloudlet *edgeproto.Cloudlet) string {
	reg := util.HostnameSanitize(*region)
	return fmt.Sprintf("%s.%s.%s.%s", ci.DnsLabel, cloudlet.DnsLabel, reg, *appDNSRoot)
}

func getAppInstFQDN(ai *edgeproto.AppInst, cloudlet *edgeproto.Cloudlet) string {
	reg := util.HostnameSanitize(*region)
	return fmt.Sprintf("%s.%s.%s.%s", ai.DnsLabel, cloudlet.DnsLabel, reg, *appDNSRoot)
}
