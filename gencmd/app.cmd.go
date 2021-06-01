// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

package gencmd

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud/cli"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
	"io"
	math "math"
	"strings"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
func AppHideTags(in *edgeproto.App) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
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
	if _, found := tags["nocmp"]; found {
		in.Revision = ""
	}
	if _, found := tags["nocmp"]; found {
		in.DeletePrepare = false
	}
	if _, found := tags["timestamp"]; found {
		in.CreatedAt = distributed_match_engine.Timestamp{}
	}
	if _, found := tags["timestamp"]; found {
		in.UpdatedAt = distributed_match_engine.Timestamp{}
	}
	for i0 := 0; i0 < len(in.RequiredOutboundConnections); i0++ {
	}
}

var AppApiCmd edgeproto.AppApiClient

var CreateAppCmd = &cli.Command{
	Use:          "CreateApp",
	RequiredArgs: strings.Join(AppRequiredArgs, " "),
	OptionalArgs: strings.Join(AppOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     AppComments,
	ReqData:      &edgeproto.App{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreateApp,
}

func runCreateApp(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.App)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return CreateApp(c, obj)
}

func CreateApp(c *cli.Command, in *edgeproto.App) error {
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
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreateApps(c *cli.Command, data []edgeproto.App, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateApp %v\n", data[ii])
		myerr := CreateApp(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteAppCmd = &cli.Command{
	Use:          "DeleteApp",
	RequiredArgs: strings.Join(AppRequiredArgs, " "),
	OptionalArgs: strings.Join(AppOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     AppComments,
	ReqData:      &edgeproto.App{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteApp,
}

func runDeleteApp(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.App)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return DeleteApp(c, obj)
}

func DeleteApp(c *cli.Command, in *edgeproto.App) error {
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
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeleteApps(c *cli.Command, data []edgeproto.App, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteApp %v\n", data[ii])
		myerr := DeleteApp(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateAppCmd = &cli.Command{
	Use:          "UpdateApp",
	RequiredArgs: strings.Join(AppRequiredArgs, " "),
	OptionalArgs: strings.Join(AppOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     AppComments,
	ReqData:      &edgeproto.App{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateApp,
}

func runUpdateApp(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.App)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData, cli.JsonNamespace)
	return UpdateApp(c, obj)
}

func UpdateApp(c *cli.Command, in *edgeproto.App) error {
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
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateApps(c *cli.Command, data []edgeproto.App, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateApp %v\n", data[ii])
		myerr := UpdateApp(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowAppCmd = &cli.Command{
	Use:          "ShowApp",
	OptionalArgs: strings.Join(append(AppRequiredArgs, AppOptionalArgs...), " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     AppComments,
	ReqData:      &edgeproto.App{},
	ReplyData:    &edgeproto.App{},
	Run:          runShowApp,
}

func runShowApp(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.App)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowApp(c, obj)
}

func ShowApp(c *cli.Command, in *edgeproto.App) error {
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
			errstr := err.Error()
			st, ok := status.FromError(err)
			if ok {
				errstr = st.Message()
			}
			return fmt.Errorf("ShowApp recv failed: %s", errstr)
		}
		AppHideTags(obj)
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowApps(c *cli.Command, data []edgeproto.App, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowApp %v\n", data[ii])
		myerr := ShowApp(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AddAppAutoProvPolicyCmd = &cli.Command{
	Use:          "AddAppAutoProvPolicy",
	RequiredArgs: strings.Join(AppAutoProvPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppAutoProvPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAutoProvPolicyAliasArgs, " "),
	SpecialArgs:  &AppAutoProvPolicySpecialArgs,
	Comments:     AppAutoProvPolicyComments,
	ReqData:      &edgeproto.AppAutoProvPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runAddAppAutoProvPolicy,
}

func runAddAppAutoProvPolicy(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AppAutoProvPolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return AddAppAutoProvPolicy(c, obj)
}

func AddAppAutoProvPolicy(c *cli.Command, in *edgeproto.AppAutoProvPolicy) error {
	if AppApiCmd == nil {
		return fmt.Errorf("AppApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AppApiCmd.AddAppAutoProvPolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("AddAppAutoProvPolicy failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func AddAppAutoProvPolicys(c *cli.Command, data []edgeproto.AppAutoProvPolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("AddAppAutoProvPolicy %v\n", data[ii])
		myerr := AddAppAutoProvPolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var RemoveAppAutoProvPolicyCmd = &cli.Command{
	Use:          "RemoveAppAutoProvPolicy",
	RequiredArgs: strings.Join(AppAutoProvPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppAutoProvPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAutoProvPolicyAliasArgs, " "),
	SpecialArgs:  &AppAutoProvPolicySpecialArgs,
	Comments:     AppAutoProvPolicyComments,
	ReqData:      &edgeproto.AppAutoProvPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRemoveAppAutoProvPolicy,
}

func runRemoveAppAutoProvPolicy(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AppAutoProvPolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return RemoveAppAutoProvPolicy(c, obj)
}

func RemoveAppAutoProvPolicy(c *cli.Command, in *edgeproto.AppAutoProvPolicy) error {
	if AppApiCmd == nil {
		return fmt.Errorf("AppApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AppApiCmd.RemoveAppAutoProvPolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("RemoveAppAutoProvPolicy failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func RemoveAppAutoProvPolicys(c *cli.Command, data []edgeproto.AppAutoProvPolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("RemoveAppAutoProvPolicy %v\n", data[ii])
		myerr := RemoveAppAutoProvPolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppApiCmds = []*cobra.Command{
	CreateAppCmd.GenCmd(),
	DeleteAppCmd.GenCmd(),
	UpdateAppCmd.GenCmd(),
	ShowAppCmd.GenCmd(),
	AddAppAutoProvPolicyCmd.GenCmd(),
	RemoveAppAutoProvPolicyCmd.GenCmd(),
}

var RemoteConnectionRequiredArgs = []string{}
var RemoteConnectionOptionalArgs = []string{
	"protocol",
	"port",
	"remoteip",
}
var RemoteConnectionAliasArgs = []string{}
var RemoteConnectionComments = map[string]string{
	"protocol": "tcp, udp or icmp",
	"port":     "TCP or UDP port",
	"remoteip": "remote IP X.X.X.X",
}
var RemoteConnectionSpecialArgs = map[string]string{}
var AppKeyRequiredArgs = []string{}
var AppKeyOptionalArgs = []string{
	"organization",
	"name",
	"version",
}
var AppKeyAliasArgs = []string{}
var AppKeyComments = map[string]string{
	"organization": "App developer organization",
	"name":         "App name",
	"version":      "App version",
}
var AppKeySpecialArgs = map[string]string{}
var ConfigFileRequiredArgs = []string{}
var ConfigFileOptionalArgs = []string{
	"kind",
	"config",
}
var ConfigFileAliasArgs = []string{}
var ConfigFileComments = map[string]string{
	"kind":   "Kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml",
	"config": "Config file contents or URI reference",
}
var ConfigFileSpecialArgs = map[string]string{}
var AppRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
}
var AppOptionalArgs = []string{
	"imagepath",
	"imagetype",
	"accessports",
	"defaultflavor",
	"authpublickey",
	"command",
	"annotations",
	"deployment",
	"deploymentmanifest",
	"deploymentgenerator",
	"androidpackagename",
	"configs:#.kind",
	"configs:#.config",
	"scalewithcluster",
	"internalports",
	"revision",
	"officialfqdn",
	"md5sum",
	"accesstype",
	"autoprovpolicies",
	"templatedelimiter",
	"skiphcports",
	"trusted",
	"requiredoutboundconnections:#.protocol",
	"requiredoutboundconnections:#.port",
	"requiredoutboundconnections:#.remoteip",
	"allowserverless",
	"serverlessconfig.vcpus",
	"serverlessconfig.ram",
	"serverlessconfig.minreplicas",
}
var AppAliasArgs = []string{
	"app-org=key.organization",
	"appname=key.name",
	"appvers=key.version",
	"defaultflavor=defaultflavor.name",
}
var AppComments = map[string]string{
	"fields":                                 "Fields are used for the Update API to specify which fields to apply",
	"app-org":                                "App developer organization",
	"appname":                                "App name",
	"appvers":                                "App version",
	"imagepath":                              "URI of where image resides",
	"imagetype":                              "Image type (see ImageType), one of ImageTypeUnknown, ImageTypeDocker, ImageTypeQcow, ImageTypeHelm, ImageTypeOvf",
	"accessports":                            "Comma separated list of protocol:port pairs that the App listens on. Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443",
	"defaultflavor":                          "Flavor name",
	"authpublickey":                          "Public key used for authentication",
	"command":                                "Command that the container runs to start service",
	"annotations":                            "Annotations is a comma separated map of arbitrary key value pairs, for example: key1=val1,key2=val2,key3=val 3",
	"deployment":                             "Deployment type (kubernetes, docker, or vm)",
	"deploymentmanifest":                     "Deployment manifest is the deployment specific manifest file/config. For docker deployment, this can be a docker-compose or docker run file. For kubernetes deployment, this can be a kubernetes yaml or helm chart file.",
	"deploymentgenerator":                    "Deployment generator target to generate a basic deployment manifest",
	"androidpackagename":                     "Android package name used to match the App name from the Android package",
	"delopt":                                 "Override actions to Controller, one of NoAutoDelete, AutoDelete",
	"configs:#.kind":                         "Kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml",
	"configs:#.config":                       "Config file contents or URI reference",
	"scalewithcluster":                       "Option to run App on all nodes of the cluster",
	"internalports":                          "Should this app have access to outside world?",
	"revision":                               "Revision can be specified or defaults to current timestamp when app is updated",
	"officialfqdn":                           "Official FQDN is the FQDN that the app uses to connect by default",
	"md5sum":                                 "MD5Sum of the VM-based app image",
	"autoprovpolicy":                         "(_deprecated_) Auto provisioning policy name",
	"accesstype":                             "Access type, one of AccessTypeDefaultForDeployment, AccessTypeDirect, AccessTypeLoadBalancer",
	"deleteprepare":                          "Preparing to be deleted",
	"autoprovpolicies":                       "Auto provisioning policy names, may be specified multiple times",
	"templatedelimiter":                      "Delimiter to be used for template parsing, defaults to [[ ]]",
	"skiphcports":                            "Comma separated list of protocol:port pairs that we should not run health check on. Should be configured in case app does not always listen on these ports. all can be specified if no health check to be run for this app. Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443.",
	"trusted":                                "Indicates that an instance of this app can be started on a trusted cloudlet",
	"requiredoutboundconnections:#.protocol": "tcp, udp or icmp",
	"requiredoutboundconnections:#.port":     "TCP or UDP port",
	"requiredoutboundconnections:#.remoteip": "remote IP X.X.X.X",
	"allowserverless":                        "App is allowed to deploy as serverless containers",
	"serverlessconfig.vcpus":                 "Virtual CPUs allocation per container when serverless, may be fractional in increments of 0.001",
	"serverlessconfig.ram":                   "RAM allocation in megabytes per container when serverless",
	"serverlessconfig.minreplicas":           "Minimum number of replicas when serverless",
}
var AppSpecialArgs = map[string]string{
	"autoprovpolicies": "StringArray",
	"fields":           "StringArray",
}
var ServerlessConfigRequiredArgs = []string{}
var ServerlessConfigOptionalArgs = []string{
	"vcpus",
	"ram",
	"minreplicas",
}
var ServerlessConfigAliasArgs = []string{}
var ServerlessConfigComments = map[string]string{
	"vcpus":       "Virtual CPUs allocation per container when serverless, may be fractional in increments of 0.001",
	"ram":         "RAM allocation in megabytes per container when serverless",
	"minreplicas": "Minimum number of replicas when serverless",
}
var ServerlessConfigSpecialArgs = map[string]string{}
var AppAutoProvPolicyRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"autoprovpolicy",
}
var AppAutoProvPolicyOptionalArgs = []string{}
var AppAutoProvPolicyAliasArgs = []string{
	"app-org=appkey.organization",
	"appname=appkey.name",
	"appvers=appkey.version",
}
var AppAutoProvPolicyComments = map[string]string{
	"app-org":        "App developer organization",
	"appname":        "App name",
	"appvers":        "App version",
	"autoprovpolicy": "Auto provisioning policy name",
}
var AppAutoProvPolicySpecialArgs = map[string]string{}
