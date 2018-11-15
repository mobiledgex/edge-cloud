// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

/*
Package gencmd is a generated protocol buffer package.

It is generated from these files:
	app.proto
	app_inst.proto
	cloud-resource-manager.proto
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

It has these top-level messages:
	AppKey
	App
	AppInstKey
	AppInst
	AppInstInfo
	AppInstMetrics
	CloudResource
	EdgeCloudApp
	EdgeCloudApplication
	CloudletKey
	Cloudlet
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
	NoticeReply
	NoticeRequest
	OperatorKey
	Operator
	CloudletRefs
	ClusterRefs
	Result
*/
package gencmd

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
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
var AppInIpAccess string
var ImageTypeStrings = []string{
	"ImageTypeUnknown",
	"ImageTypeDocker",
	"ImageTypeQCOW",
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
func AppSlicer(in *edgeproto.App) []string {
	s := make([]string, 0, 12)
	if in.Fields == nil {
		in.Fields = make([]string, 1)
	}
	s = append(s, in.Fields[0])
	s = append(s, in.Key.DeveloperKey.Name)
	s = append(s, in.Key.Name)
	s = append(s, in.Key.Version)
	s = append(s, in.ImagePath)
	s = append(s, edgeproto.ImageType_name[int32(in.ImageType)])
	s = append(s, edgeproto.IpAccess_name[int32(in.IpAccess)])
	s = append(s, in.AccessPorts)
	s = append(s, in.Config)
	s = append(s, in.DefaultFlavor.Name)
	s = append(s, in.Cluster.Name)
	s = append(s, in.AppTemplate)
	s = append(s, in.AuthPublicKey)
	s = append(s, in.AndroidPackageName)
	return s
}

func AppHeaderSlicer() []string {
	s := make([]string, 0, 12)
	s = append(s, "Fields")
	s = append(s, "Key-DeveloperKey-Name")
	s = append(s, "Key-Name")
	s = append(s, "Key-Version")
	s = append(s, "ImagePath")
	s = append(s, "ImageType")
	s = append(s, "IpAccess")
	s = append(s, "AccessPorts")
	s = append(s, "Config")
	s = append(s, "DefaultFlavor-Name")
	s = append(s, "Cluster-Name")
	s = append(s, "AppTemplate")
	s = append(s, "AuthPublicKey")
	s = append(s, "AndroidPackageName")
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
	AppFlagSet.StringVar(&AppInIpAccess, "ipaccess", "", "one of [IpAccessUnknown IpAccessDedicated IpAccessDedicatedOrShared IpAccessShared]")
	AppFlagSet.StringVar(&AppIn.AccessPorts, "accessports", "", "AccessPorts")
	AppFlagSet.StringVar(&AppIn.Config, "config", "", "Config")
	AppFlagSet.StringVar(&AppIn.DefaultFlavor.Name, "defaultflavor-name", "", "DefaultFlavor.Name")
	AppFlagSet.StringVar(&AppIn.Cluster.Name, "cluster-name", "", "Cluster.Name")
	AppFlagSet.StringVar(&AppIn.AppTemplate, "apptemplate", "", "AppTemplate")
	AppFlagSet.StringVar(&AppIn.AuthPublicKey, "authpublickey", "", "AuthPublicKey")
	AppFlagSet.StringVar(&AppIn.AndroidPackageName, "androidpackagename", "", "AndroidPackageName")
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
	if AppFlagSet.Lookup("ipaccess").Changed {
		AppIn.Fields = append(AppIn.Fields, "6")
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
	if AppFlagSet.Lookup("androidpackagename").Changed {
		AppIn.Fields = append(AppIn.Fields, "13")
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
	if AppInIpAccess != "" {
		switch AppInIpAccess {
		case "IpAccessUnknown":
			AppIn.IpAccess = edgeproto.IpAccess(0)
		case "IpAccessDedicated":
			AppIn.IpAccess = edgeproto.IpAccess(1)
		case "IpAccessDedicatedOrShared":
			AppIn.IpAccess = edgeproto.IpAccess(2)
		case "IpAccessShared":
			AppIn.IpAccess = edgeproto.IpAccess(3)
		default:
			return errors.New("Invalid value for AppInIpAccess")
		}
	}
	return nil
}
