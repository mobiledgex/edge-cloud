// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

/*
Package gencmd is a generated protocol buffer package.

It is generated from these files:
	app.proto
	app_inst.proto
	cloud-resource-manager.proto
	cloudlet.proto
	developer.proto
	notice.proto
	operator.proto
	result.proto

It has these top-level messages:
	AppKey
	App
	AppInstKey
	AppInst
	CloudResource
	EdgeCloudApp
	EdgeCloudApplication
	CloudletKey
	Cloudlet
	DeveloperKey
	Developer
	NoticeReply
	NoticeRequest
	OperatorCode
	OperatorKey
	Operator
	Result
*/
package gencmd

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "time"
import "github.com/spf13/cobra"
import "context"
import "os"
import "io"
import "text/tabwriter"
import "github.com/spf13/pflag"
import "encoding/json"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/cmdsup"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var AppApiCmd edgeproto.AppApiClient
var AppIn edgeproto.App
var AppFlagSet = pflag.NewFlagSet("App", pflag.ExitOnError)

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

func AppSlicer(in *edgeproto.App) []string {
	s := make([]string, 0, 3)
	if in.Fields == nil {
		in.Fields = make([]string, 1)
	}
	s = append(s, in.Fields[0])
	s = append(s, in.Key.DeveloperKey.Name)
	s = append(s, in.Key.Name)
	s = append(s, in.Key.Version)
	s = append(s, in.AppPath)
	return s
}

func AppHeaderSlicer() []string {
	s := make([]string, 0, 3)
	s = append(s, "Fields")
	s = append(s, "Key-DeveloperKey-Name")
	s = append(s, "Key-Name")
	s = append(s, "Key-Version")
	s = append(s, "AppPath")
	return s
}

var CreateAppCmd = &cobra.Command{
	Use: "CreateApp",
	Run: func(cmd *cobra.Command, args []string) {
		if AppApiCmd == nil {
			fmt.Println("AppApi client not initialized")
			return
		}
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		objs, err := AppApiCmd.CreateApp(ctx, &AppIn)
		cancel()
		if err != nil {
			fmt.Println("CreateApp failed: ", err)
			return
		}
		switch cmdsup.OutputFormat {
		case cmdsup.OutputFormatYaml:
			output, err := yaml.Marshal(objs)
			if err != nil {
				fmt.Printf("Yaml failed to marshal: %s\n", err)
				return
			}
			fmt.Print(string(output))
		case cmdsup.OutputFormatJson:
			output, err := json.MarshalIndent(objs, "", "  ")
			if err != nil {
				fmt.Printf("Json failed to marshal: %s\n", err)
				return
			}
			fmt.Println(string(output))
		case cmdsup.OutputFormatJsonCompact:
			output, err := json.Marshal(objs)
			if err != nil {
				fmt.Printf("Json failed to marshal: %s\n", err)
				return
			}
			fmt.Println(string(output))
		case cmdsup.OutputFormatTable:
			output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(output, strings.Join(ResultHeaderSlicer(), "\t"))
			fmt.Fprintln(output, strings.Join(ResultSlicer(objs), "\t"))
			output.Flush()
		}
	},
}

var DeleteAppCmd = &cobra.Command{
	Use: "DeleteApp",
	Run: func(cmd *cobra.Command, args []string) {
		if AppApiCmd == nil {
			fmt.Println("AppApi client not initialized")
			return
		}
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		objs, err := AppApiCmd.DeleteApp(ctx, &AppIn)
		cancel()
		if err != nil {
			fmt.Println("DeleteApp failed: ", err)
			return
		}
		switch cmdsup.OutputFormat {
		case cmdsup.OutputFormatYaml:
			output, err := yaml.Marshal(objs)
			if err != nil {
				fmt.Printf("Yaml failed to marshal: %s\n", err)
				return
			}
			fmt.Print(string(output))
		case cmdsup.OutputFormatJson:
			output, err := json.MarshalIndent(objs, "", "  ")
			if err != nil {
				fmt.Printf("Json failed to marshal: %s\n", err)
				return
			}
			fmt.Println(string(output))
		case cmdsup.OutputFormatJsonCompact:
			output, err := json.Marshal(objs)
			if err != nil {
				fmt.Printf("Json failed to marshal: %s\n", err)
				return
			}
			fmt.Println(string(output))
		case cmdsup.OutputFormatTable:
			output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(output, strings.Join(ResultHeaderSlicer(), "\t"))
			fmt.Fprintln(output, strings.Join(ResultSlicer(objs), "\t"))
			output.Flush()
		}
	},
}

var UpdateAppCmd = &cobra.Command{
	Use: "UpdateApp",
	Run: func(cmd *cobra.Command, args []string) {
		if AppApiCmd == nil {
			fmt.Println("AppApi client not initialized")
			return
		}
		var err error
		AppSetFields()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		objs, err := AppApiCmd.UpdateApp(ctx, &AppIn)
		cancel()
		if err != nil {
			fmt.Println("UpdateApp failed: ", err)
			return
		}
		switch cmdsup.OutputFormat {
		case cmdsup.OutputFormatYaml:
			output, err := yaml.Marshal(objs)
			if err != nil {
				fmt.Printf("Yaml failed to marshal: %s\n", err)
				return
			}
			fmt.Print(string(output))
		case cmdsup.OutputFormatJson:
			output, err := json.MarshalIndent(objs, "", "  ")
			if err != nil {
				fmt.Printf("Json failed to marshal: %s\n", err)
				return
			}
			fmt.Println(string(output))
		case cmdsup.OutputFormatJsonCompact:
			output, err := json.Marshal(objs)
			if err != nil {
				fmt.Printf("Json failed to marshal: %s\n", err)
				return
			}
			fmt.Println(string(output))
		case cmdsup.OutputFormatTable:
			output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(output, strings.Join(ResultHeaderSlicer(), "\t"))
			fmt.Fprintln(output, strings.Join(ResultSlicer(objs), "\t"))
			output.Flush()
		}
	},
}

var ShowAppCmd = &cobra.Command{
	Use: "ShowApp",
	Run: func(cmd *cobra.Command, args []string) {
		if AppApiCmd == nil {
			fmt.Println("AppApi client not initialized")
			return
		}
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		stream, err := AppApiCmd.ShowApp(ctx, &AppIn)
		if err != nil {
			fmt.Println("ShowApp failed: ", err)
			return
		}
		objs := make([]*edgeproto.App, 0)
		for {
			obj, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("ShowApp recv failed: ", err)
				break
			}
			objs = append(objs, obj)
		}
		if len(objs) == 0 {
			return
		}
		switch cmdsup.OutputFormat {
		case cmdsup.OutputFormatYaml:
			output, err := yaml.Marshal(objs)
			if err != nil {
				fmt.Printf("Yaml failed to marshal: %s\n", err)
				return
			}
			fmt.Print(string(output))
		case cmdsup.OutputFormatJson:
			output, err := json.MarshalIndent(objs, "", "  ")
			if err != nil {
				fmt.Printf("Json failed to marshal: %s\n", err)
				return
			}
			fmt.Println(string(output))
		case cmdsup.OutputFormatJsonCompact:
			output, err := json.Marshal(objs)
			if err != nil {
				fmt.Printf("Json failed to marshal: %s\n", err)
				return
			}
			fmt.Println(string(output))
		case cmdsup.OutputFormatTable:
			output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(output, strings.Join(AppHeaderSlicer(), "\t"))
			for _, obj := range objs {
				fmt.Fprintln(output, strings.Join(AppSlicer(obj), "\t"))
			}
			output.Flush()
		}
	},
}

func init() {
	AppFlagSet.StringVar(&AppIn.Key.DeveloperKey.Name, "key-developerkey-name", "", "Key.DeveloperKey.Name")
	AppFlagSet.StringVar(&AppIn.Key.Name, "key-name", "", "Key.Name")
	AppFlagSet.StringVar(&AppIn.Key.Version, "key-version", "", "Key.Version")
	AppFlagSet.StringVar(&AppIn.AppPath, "apppath", "", "AppPath")
	CreateAppCmd.Flags().AddFlagSet(AppFlagSet)
	DeleteAppCmd.Flags().AddFlagSet(AppFlagSet)
	UpdateAppCmd.Flags().AddFlagSet(AppFlagSet)
	ShowAppCmd.Flags().AddFlagSet(AppFlagSet)
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
	if AppFlagSet.Lookup("apppath").Changed {
		AppIn.Fields = append(AppIn.Fields, "4")
	}
}
