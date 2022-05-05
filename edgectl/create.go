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

	"github.com/edgexr/edge-cloud/cli"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/gencmd"
)

var createCmd = &cli.Command{
	Use:          "Create",
	DataFlagOnly: true,
	ReqData:      &edgeproto.AllData{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreate,
}

func runCreate(c *cli.Command, args []string) error {
	mapped, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	data := c.ReqData.(*edgeproto.AllData)

	gencmd.CreateFlavors(c, data.Flavors, &err)
	if data.Settings != nil {
		objMap, err := cli.GetGenericObj(mapped.Data["settings"])
		if err != nil {
			return fmt.Errorf("invalid data map for settings: %v", err)
		}
		settingsMap := &cli.MapData{
			Namespace: mapped.Namespace,
			Data:      objMap,
		}
		data.Settings.Fields = cli.GetSpecifiedFields(settingsMap, data.Settings)
		gencmd.UpdateSettingsBatch(c, data.Settings, &err)
	}

	gencmd.CreateOperatorCodes(c, data.OperatorCodes, &err)
	gencmd.CreateCloudlets(c, data.Cloudlets, &err)
	gencmd.CreateAutoScalePolicys(c, data.AutoScalePolicies, &err)
	gencmd.CreateAutoProvPolicys(c, data.AutoProvPolicies, &err)
	gencmd.CreateApps(c, data.Apps, &err)
	gencmd.CreateTrustPolicys(c, data.TrustPolicies, &err)
	gencmd.CreateTrustPolicyExceptions(c, data.TrustPolicyExceptions, &err)
	gencmd.CreateNetworks(c, data.Networks, &err)
	gencmd.CreateClusterInsts(c, data.ClusterInsts, &err)
	gencmd.CreateAppInsts(c, data.AppInstances, &err)
	gencmd.CreateAlertPolicys(c, data.AlertPolicies, &err)
	return err
}

var deleteCmd = &cli.Command{
	Use:          "Delete",
	DataFlagOnly: true,
	ReqData:      &edgeproto.AllData{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDelete,
}

func runDelete(c *cli.Command, args []string) error {
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	data := c.ReqData.(*edgeproto.AllData)

	gencmd.DeleteAppInsts(c, data.AppInstances, &err)
	gencmd.DeleteClusterInsts(c, data.ClusterInsts, &err)
	gencmd.DeleteTrustPolicys(c, data.TrustPolicies, &err)
	gencmd.DeleteTrustPolicyExceptions(c, data.TrustPolicyExceptions, &err)
	gencmd.DeleteNetworks(c, data.Networks, &err)
	gencmd.DeleteApps(c, data.Apps, &err)
	gencmd.DeleteAutoProvPolicys(c, data.AutoProvPolicies, &err)
	gencmd.DeleteAutoScalePolicys(c, data.AutoScalePolicies, &err)
	gencmd.DeleteCloudlets(c, data.Cloudlets, &err)
	gencmd.DeleteOperatorCodes(c, data.OperatorCodes, &err)

	if data.Settings != nil {
		gencmd.ResetSettingsBatch(c, data.Settings, &err)
	}
	gencmd.DeleteFlavors(c, data.Flavors, &err)
	gencmd.DeleteAlertPolicys(c, data.AlertPolicies, &err)
	return err
}
