// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

/*
Package gencmd is a generated protocol buffer package.

It is generated from these files:
	app.proto
	app_inst.proto
	cloudlet.proto
	cluster.proto
	clusterflavor.proto
	clusterinst.proto
	common.proto
	controller.proto
	developer.proto
	flavor.proto
	metric.proto
	node.proto
	notice.proto
	operator.proto
	refs.proto
	result.proto
	version.proto

It has these top-level messages:
	AppKey
	ConfigFile
	App
	AppInstKey
	AppInst
	AppInstInfo
	AppInstMetrics
	CloudletKey
	CloudletInfraCommon
	AzureProperties
	GcpProperties
	OpenStackProperties
	CloudletInfraProperties
	Cloudlet
	FlavorInfo
	CloudletInfo
	CloudletMetrics
	ClusterKey
	Cluster
	ClusterFlavorKey
	ClusterFlavor
	ClusterInstKey
	ClusterInst
	ClusterInstInfo
	ControllerKey
	Controller
	DeveloperKey
	Developer
	FlavorKey
	Flavor
	MetricTag
	MetricVal
	Metric
	NodeKey
	Node
	Notice
	OperatorKey
	Operator
	CloudletRefs
	ClusterRefs
	Result
	DataModelVersion
*/
package gencmd

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "strconv"
import "github.com/spf13/cobra"
import "context"
import "os"
import "io"
import "text/tabwriter"
import "github.com/spf13/pflag"
import "errors"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/cmdsup"
import "google.golang.org/grpc/status"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var AppApiCmd edgeproto.AppApiClient
var AppIn edgeproto.App
var AppFlagSet = pflag.NewFlagSet("App", pflag.ExitOnError)
var AppNoConfigFlagSet = pflag.NewFlagSet("AppNoConfig", pflag.ExitOnError)
var AppInImageType string
var AppInDelOpt string
var ImageTypeStrings = []string{
	"ImageTypeUnknown",
	"ImageTypeDocker",
	"ImageTypeQCOW",
}

var DeleteTypeStrings = []string{
	"NoAutoDelete",
	"AutoDelete",
}

func AppKeySlicer(in *edgeproto.AppKey) []string {
	s := make([]string, 0, 3)
	s = append(s, in.DeveloperKey.Name)
	s = append(s, in.Name)
	s = append(s, in.Version)
	return s
}

func AppKeyHeaderSlicer() []string {
	s := make([]string, 0, 3)
	s = append(s, "DeveloperKey-Name")
	s = append(s, "Name")
	s = append(s, "Version")
	return s
}

func AppKeyWriteOutputArray(objs []*edgeproto.AppKey) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppKeyHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(AppKeySlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func AppKeyWriteOutputOne(obj *edgeproto.AppKey) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppKeyHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(AppKeySlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func ConfigFileSlicer(in *edgeproto.ConfigFile) []string {
	s := make([]string, 0, 2)
	s = append(s, in.Kind)
	s = append(s, in.Config)
	return s
}

func ConfigFileHeaderSlicer() []string {
	s := make([]string, 0, 2)
	s = append(s, "Kind")
	s = append(s, "Config")
	return s
}

func ConfigFileWriteOutputArray(objs []*edgeproto.ConfigFile) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(ConfigFileHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(ConfigFileSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func ConfigFileWriteOutputOne(obj *edgeproto.ConfigFile) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(ConfigFileHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(ConfigFileSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func AppSlicer(in *edgeproto.App) []string {
	s := make([]string, 0, 19)
	if in.Fields == nil {
		in.Fields = make([]string, 1)
	}
	s = append(s, in.Fields[0])
	s = append(s, in.Key.DeveloperKey.Name)
	s = append(s, in.Key.Name)
	s = append(s, in.Key.Version)
	s = append(s, in.ImagePath)
	s = append(s, edgeproto.ImageType_name[int32(in.ImageType)])
	s = append(s, in.AccessPorts)
	s = append(s, in.Config)
	s = append(s, in.DefaultFlavor.Name)
	s = append(s, in.Cluster.Name)
	s = append(s, in.AppTemplate)
	s = append(s, in.AuthPublicKey)
	s = append(s, in.Command)
	s = append(s, in.Annotations)
	s = append(s, in.Deployment)
	s = append(s, in.DeploymentManifest)
	s = append(s, in.DeploymentGenerator)
	s = append(s, in.AndroidPackageName)
	s = append(s, strconv.FormatBool(in.PermitsPlatformApps))
	s = append(s, edgeproto.DeleteType_name[int32(in.DelOpt)])
	if in.Configs == nil {
		in.Configs = make([]*edgeproto.ConfigFile, 1)
	}
	if in.Configs[0] == nil {
		in.Configs[0] = &edgeproto.ConfigFile{}
	}
	s = append(s, in.Configs[0].Kind)
	s = append(s, in.Configs[0].Config)
	return s
}

func AppHeaderSlicer() []string {
	s := make([]string, 0, 19)
	s = append(s, "Fields")
	s = append(s, "Key-DeveloperKey-Name")
	s = append(s, "Key-Name")
	s = append(s, "Key-Version")
	s = append(s, "ImagePath")
	s = append(s, "ImageType")
	s = append(s, "AccessPorts")
	s = append(s, "Config")
	s = append(s, "DefaultFlavor-Name")
	s = append(s, "Cluster-Name")
	s = append(s, "AppTemplate")
	s = append(s, "AuthPublicKey")
	s = append(s, "Command")
	s = append(s, "Annotations")
	s = append(s, "Deployment")
	s = append(s, "DeploymentManifest")
	s = append(s, "DeploymentGenerator")
	s = append(s, "AndroidPackageName")
	s = append(s, "PermitsPlatformApps")
	s = append(s, "DelOpt")
	s = append(s, "Configs-Kind")
	s = append(s, "Configs-Config")
	return s
}

func AppWriteOutputArray(objs []*edgeproto.App) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(AppSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func AppWriteOutputOne(obj *edgeproto.App) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(AppSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func AppHideTags(in *edgeproto.App) {
	if cmdsup.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cmdsup.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.DeploymentManifest = ""
	}
	if _, found := tags["nocmp"]; found {
		in.DeploymentGenerator = ""
	}
	if _, found := tags["nocmp"]; found {
		in.DelOpt = 0
	}
	for i0 := 0; i0 < len(in.Configs); i0++ {
	}
}

var CreateAppCmd = &cobra.Command{
	Use: "CreateApp",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseAppEnums()
		if err != nil {
			return fmt.Errorf("CreateApp failed: %s", err.Error())
		}
		return CreateApp(&AppIn)
	},
}

func CreateApp(in *edgeproto.App) error {
	if AppApiCmd == nil {
		return fmt.Errorf("AppApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AppApiCmd.CreateApp(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateApp failed: %s", errstr)
	}
	ResultWriteOutputOne(obj)
	return nil
}

func CreateApps(data []edgeproto.App, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateApp %v\n", data[ii])
		myerr := CreateApp(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteAppCmd = &cobra.Command{
	Use: "DeleteApp",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseAppEnums()
		if err != nil {
			return fmt.Errorf("DeleteApp failed: %s", err.Error())
		}
		return DeleteApp(&AppIn)
	},
}

func DeleteApp(in *edgeproto.App) error {
	if AppApiCmd == nil {
		return fmt.Errorf("AppApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AppApiCmd.DeleteApp(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeleteApp failed: %s", errstr)
	}
	ResultWriteOutputOne(obj)
	return nil
}

func DeleteApps(data []edgeproto.App, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteApp %v\n", data[ii])
		myerr := DeleteApp(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateAppCmd = &cobra.Command{
	Use: "UpdateApp",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseAppEnums()
		if err != nil {
			return fmt.Errorf("UpdateApp failed: %s", err.Error())
		}
		AppSetFields()
		return UpdateApp(&AppIn)
	},
}

func UpdateApp(in *edgeproto.App) error {
	if AppApiCmd == nil {
		return fmt.Errorf("AppApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AppApiCmd.UpdateApp(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateApp failed: %s", errstr)
	}
	ResultWriteOutputOne(obj)
	return nil
}

func UpdateApps(data []edgeproto.App, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateApp %v\n", data[ii])
		myerr := UpdateApp(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowAppCmd = &cobra.Command{
	Use: "ShowApp",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseAppEnums()
		if err != nil {
			return fmt.Errorf("ShowApp failed: %s", err.Error())
		}
		return ShowApp(&AppIn)
	},
}

func ShowApp(in *edgeproto.App) error {
	if AppApiCmd == nil {
		return fmt.Errorf("AppApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AppApiCmd.ShowApp(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowApp failed: %s", errstr)
	}
	objs := make([]*edgeproto.App, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ShowApp recv failed: %s", err.Error())
		}
		AppHideTags(obj)
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	AppWriteOutputArray(objs)
	return nil
}

func ShowApps(data []edgeproto.App, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowApp %v\n", data[ii])
		myerr := ShowApp(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppApiCmds = []*cobra.Command{
	CreateAppCmd,
	DeleteAppCmd,
	UpdateAppCmd,
	ShowAppCmd,
}

func init() {
	AppFlagSet.StringVar(&AppIn.Key.DeveloperKey.Name, "key-developerkey-name", "", "Key.DeveloperKey.Name")
	AppFlagSet.StringVar(&AppIn.Key.Name, "key-name", "", "Key.Name")
	AppFlagSet.StringVar(&AppIn.Key.Version, "key-version", "", "Key.Version")
	AppFlagSet.StringVar(&AppIn.ImagePath, "imagepath", "", "ImagePath")
	AppFlagSet.StringVar(&AppInImageType, "imagetype", "", "one of [ImageTypeUnknown ImageTypeDocker ImageTypeQCOW]")
	AppFlagSet.StringVar(&AppIn.AccessPorts, "accessports", "", "AccessPorts")
	AppFlagSet.StringVar(&AppIn.Config, "config", "", "Config")
	AppFlagSet.StringVar(&AppIn.DefaultFlavor.Name, "defaultflavor-name", "", "DefaultFlavor.Name")
	AppFlagSet.StringVar(&AppIn.Cluster.Name, "cluster-name", "", "Cluster.Name")
	AppFlagSet.StringVar(&AppIn.AppTemplate, "apptemplate", "", "AppTemplate")
	AppFlagSet.StringVar(&AppIn.AuthPublicKey, "authpublickey", "", "AuthPublicKey")
	AppFlagSet.StringVar(&AppIn.Command, "command", "", "Command")
	AppFlagSet.StringVar(&AppIn.Annotations, "annotations", "", "Annotations")
	AppFlagSet.StringVar(&AppIn.Deployment, "deployment", "", "Deployment")
	AppFlagSet.StringVar(&AppIn.DeploymentManifest, "deploymentmanifest", "", "DeploymentManifest")
	AppFlagSet.StringVar(&AppIn.DeploymentGenerator, "deploymentgenerator", "", "DeploymentGenerator")
	AppFlagSet.StringVar(&AppIn.AndroidPackageName, "androidpackagename", "", "AndroidPackageName")
	AppFlagSet.BoolVar(&AppIn.PermitsPlatformApps, "permitsplatformapps", false, "PermitsPlatformApps")
	AppFlagSet.StringVar(&AppInDelOpt, "delopt", "", "one of [NoAutoDelete AutoDelete]")
	CreateAppCmd.Flags().AddFlagSet(AppFlagSet)
	DeleteAppCmd.Flags().AddFlagSet(AppFlagSet)
	UpdateAppCmd.Flags().AddFlagSet(AppFlagSet)
	ShowAppCmd.Flags().AddFlagSet(AppFlagSet)
}

func AppApiAllowNoConfig() {
	CreateAppCmd.Flags().AddFlagSet(AppNoConfigFlagSet)
	DeleteAppCmd.Flags().AddFlagSet(AppNoConfigFlagSet)
	UpdateAppCmd.Flags().AddFlagSet(AppNoConfigFlagSet)
	ShowAppCmd.Flags().AddFlagSet(AppNoConfigFlagSet)
}

func AppSetFields() {
	AppIn.Fields = make([]string, 0)
	if AppFlagSet.Lookup("key-developerkey-name").Changed {
		AppIn.Fields = append(AppIn.Fields, "2.1.2")
	}
	if AppFlagSet.Lookup("key-name").Changed {
		AppIn.Fields = append(AppIn.Fields, "2.2")
	}
	if AppFlagSet.Lookup("key-version").Changed {
		AppIn.Fields = append(AppIn.Fields, "2.3")
	}
	if AppFlagSet.Lookup("imagepath").Changed {
		AppIn.Fields = append(AppIn.Fields, "4")
	}
	if AppFlagSet.Lookup("imagetype").Changed {
		AppIn.Fields = append(AppIn.Fields, "5")
	}
	if AppFlagSet.Lookup("accessports").Changed {
		AppIn.Fields = append(AppIn.Fields, "7")
	}
	if AppFlagSet.Lookup("config").Changed {
		AppIn.Fields = append(AppIn.Fields, "8")
	}
	if AppFlagSet.Lookup("defaultflavor-name").Changed {
		AppIn.Fields = append(AppIn.Fields, "9.1")
	}
	if AppFlagSet.Lookup("cluster-name").Changed {
		AppIn.Fields = append(AppIn.Fields, "10.1")
	}
	if AppFlagSet.Lookup("apptemplate").Changed {
		AppIn.Fields = append(AppIn.Fields, "11")
	}
	if AppFlagSet.Lookup("authpublickey").Changed {
		AppIn.Fields = append(AppIn.Fields, "12")
	}
	if AppFlagSet.Lookup("command").Changed {
		AppIn.Fields = append(AppIn.Fields, "13")
	}
	if AppFlagSet.Lookup("annotations").Changed {
		AppIn.Fields = append(AppIn.Fields, "14")
	}
	if AppFlagSet.Lookup("deployment").Changed {
		AppIn.Fields = append(AppIn.Fields, "15")
	}
	if AppFlagSet.Lookup("deploymentmanifest").Changed {
		AppIn.Fields = append(AppIn.Fields, "16")
	}
	if AppFlagSet.Lookup("deploymentgenerator").Changed {
		AppIn.Fields = append(AppIn.Fields, "17")
	}
	if AppFlagSet.Lookup("androidpackagename").Changed {
		AppIn.Fields = append(AppIn.Fields, "18")
	}
	if AppFlagSet.Lookup("permitsplatformapps").Changed {
		AppIn.Fields = append(AppIn.Fields, "19")
	}
	if AppFlagSet.Lookup("delopt").Changed {
		AppIn.Fields = append(AppIn.Fields, "20")
	}
}

func parseAppEnums() error {
	if AppInImageType != "" {
		switch AppInImageType {
		case "ImageTypeUnknown":
			AppIn.ImageType = edgeproto.ImageType(0)
		case "ImageTypeDocker":
			AppIn.ImageType = edgeproto.ImageType(1)
		case "ImageTypeQCOW":
			AppIn.ImageType = edgeproto.ImageType(2)
		default:
			return errors.New("Invalid value for AppInImageType")
		}
	}
	if AppInDelOpt != "" {
		switch AppInDelOpt {
		case "NoAutoDelete":
			AppIn.DelOpt = edgeproto.DeleteType(0)
		case "AutoDelete":
			AppIn.DelOpt = edgeproto.DeleteType(1)
		default:
			return errors.New("Invalid value for AppInDelOpt")
		}
	}
	return nil
}
