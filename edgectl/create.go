package main

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/gencmd"
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
