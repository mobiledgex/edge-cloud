package main

import (
	"io/ioutil"
	"os"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/gencmd"
	yaml "github.com/mobiledgex/yaml/v2"
	"github.com/spf13/cobra"
)

var filename string

func init() {
	createCmd.Flags().StringVarP(&filename, "filename", "f", "", "YAML filename")
	deleteCmd.Flags().StringVarP(&filename, "filename", "f", "", "YAML filename")
}

func readData(data *edgeproto.ApplicationData) error {
	input := os.Stdin
	var err error
	if filename != "" {
		input, err = os.Open(filename)
		if err != nil {
			return err
		}
	}
	in, err := ioutil.ReadAll(input)
	if err != nil {
		return err
	}
	return yaml.UnmarshalStrict(in, data)
}

var createCmd = &cobra.Command{
	Use: "Create",
	RunE: func(cmd *cobra.Command, args []string) error {
		data := edgeproto.ApplicationData{}
		err := readData(&data)
		if err != nil {
			return err
		}
		gencmd.CreateFlavors(data.Flavors, &err)
		gencmd.CreateClusterFlavors(data.ClusterFlavors, &err)
		gencmd.CreateOperators(data.Operators, &err)
		gencmd.CreateDevelopers(data.Developers, &err)
		gencmd.CreateCloudlets(data.Cloudlets, &err)
		gencmd.InjectCloudletInfos(data.CloudletInfos, &err)
		gencmd.CreateClusters(data.Clusters, &err)
		gencmd.CreateApps(data.Applications, &err)
		gencmd.CreateClusterInsts(data.ClusterInsts, &err)
		gencmd.CreateAppInsts(data.AppInstances, &err)
		return err
	},
}

var deleteCmd = &cobra.Command{
	Use: "Delete",
	RunE: func(cmd *cobra.Command, args []string) error {
		data := edgeproto.ApplicationData{}
		err := readData(&data)
		if err != nil {
			return err
		}
		gencmd.DeleteAppInsts(data.AppInstances, &err)
		gencmd.DeleteClusterInsts(data.ClusterInsts, &err)
		gencmd.DeleteApps(data.Applications, &err)
		gencmd.DeleteClusters(data.Clusters, &err)
		gencmd.EvictCloudletInfos(data.CloudletInfos, &err)
		gencmd.DeleteCloudlets(data.Cloudlets, &err)
		gencmd.DeleteDevelopers(data.Developers, &err)
		gencmd.DeleteOperators(data.Operators, &err)
		gencmd.DeleteClusterFlavors(data.ClusterFlavors, &err)
		gencmd.DeleteFlavors(data.Flavors, &err)
		return err
	},
}
