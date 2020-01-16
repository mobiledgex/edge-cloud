package main

import (
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/gencmd"
)

var createCmd = &cli.Command{
	Use:          "Create",
	DataFlagOnly: true,
	ReqData:      &edgeproto.ApplicationData{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreate,
}

func runCreate(c *cli.Command, args []string) error {
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	data := c.ReqData.(*edgeproto.ApplicationData)

	gencmd.CreateFlavors(c, data.Flavors, &err)
	gencmd.CreateOperators(c, data.Operators, &err)
	gencmd.CreateOperatorCodes(c, data.OperatorCodes, &err)
	gencmd.CreateDevelopers(c, data.Developers, &err)
	gencmd.CreateCloudlets(c, data.Cloudlets, &err)
	gencmd.InjectCloudletInfos(c, data.CloudletInfos, &err)
	gencmd.CreateAutoScalePolicys(c, data.AutoScalePolicies, &err)
	gencmd.CreateAutoProvPolicys(c, data.AutoProvPolicies, &err)
	gencmd.CreateApps(c, data.Applications, &err)
	gencmd.CreatePrivacyPolicys(c, data.PrivacyPolicies, &err)
	gencmd.CreateClusterInsts(c, data.ClusterInsts, &err)
	gencmd.CreateAppInsts(c, data.AppInstances, &err)

	return err
}

var deleteCmd = &cli.Command{
	Use:          "Delete",
	DataFlagOnly: true,
	ReqData:      &edgeproto.ApplicationData{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDelete,
}

func runDelete(c *cli.Command, args []string) error {
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	data := c.ReqData.(*edgeproto.ApplicationData)

	gencmd.DeleteAppInsts(c, data.AppInstances, &err)
	gencmd.DeleteClusterInsts(c, data.ClusterInsts, &err)
	gencmd.DeletePrivacyPolicys(c, data.PrivacyPolicies, &err)
	gencmd.DeleteApps(c, data.Applications, &err)
	gencmd.DeleteAutoProvPolicys(c, data.AutoProvPolicies, &err)
	gencmd.DeleteAutoScalePolicys(c, data.AutoScalePolicies, &err)
	gencmd.EvictCloudletInfos(c, data.CloudletInfos, &err)
	gencmd.DeleteCloudlets(c, data.Cloudlets, &err)
	gencmd.DeleteDevelopers(c, data.Developers, &err)
	gencmd.DeleteOperatorCodes(c, data.OperatorCodes, &err)
	gencmd.DeleteOperators(c, data.Operators, &err)
	gencmd.DeleteFlavors(c, data.Flavors, &err)
	return err
}
