// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: operator.proto

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
var OperatorApiCmd edgeproto.OperatorApiClient
var OperatorIn edgeproto.Operator
var OperatorFlagSet = pflag.NewFlagSet("Operator", pflag.ExitOnError)

func OperatorCodeSlicer(in *edgeproto.OperatorCode) []string {
	s := make([]string, 0, 2)
	s = append(s, in.MNC)
	s = append(s, in.MCC)
	return s
}

func OperatorCodeHeaderSlicer() []string {
	s := make([]string, 0, 2)
	s = append(s, "MNC")
	s = append(s, "MCC")
	return s
}

func OperatorKeySlicer(in *edgeproto.OperatorKey) []string {
	s := make([]string, 0, 1)
	s = append(s, in.Name)
	return s
}

func OperatorKeyHeaderSlicer() []string {
	s := make([]string, 0, 1)
	s = append(s, "Name")
	return s
}

func OperatorSlicer(in *edgeproto.Operator) []string {
	s := make([]string, 0, 2)
	if in.Fields == nil {
		in.Fields = make([]string, 1)
	}
	s = append(s, in.Fields[0])
	s = append(s, in.Key.Name)
	return s
}

func OperatorHeaderSlicer() []string {
	s := make([]string, 0, 2)
	s = append(s, "Fields")
	s = append(s, "Key-Name")
	return s
}

var CreateOperatorCmd = &cobra.Command{
	Use: "CreateOperator",
	Run: func(cmd *cobra.Command, args []string) {
		if OperatorApiCmd == nil {
			fmt.Println("OperatorApi client not initialized")
			return
		}
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		objs, err := OperatorApiCmd.CreateOperator(ctx, &OperatorIn)
		cancel()
		if err != nil {
			fmt.Println("CreateOperator failed: ", err)
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

var DeleteOperatorCmd = &cobra.Command{
	Use: "DeleteOperator",
	Run: func(cmd *cobra.Command, args []string) {
		if OperatorApiCmd == nil {
			fmt.Println("OperatorApi client not initialized")
			return
		}
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		objs, err := OperatorApiCmd.DeleteOperator(ctx, &OperatorIn)
		cancel()
		if err != nil {
			fmt.Println("DeleteOperator failed: ", err)
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

var UpdateOperatorCmd = &cobra.Command{
	Use: "UpdateOperator",
	Run: func(cmd *cobra.Command, args []string) {
		if OperatorApiCmd == nil {
			fmt.Println("OperatorApi client not initialized")
			return
		}
		var err error
		OperatorSetFields()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		objs, err := OperatorApiCmd.UpdateOperator(ctx, &OperatorIn)
		cancel()
		if err != nil {
			fmt.Println("UpdateOperator failed: ", err)
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

var ShowOperatorCmd = &cobra.Command{
	Use: "ShowOperator",
	Run: func(cmd *cobra.Command, args []string) {
		if OperatorApiCmd == nil {
			fmt.Println("OperatorApi client not initialized")
			return
		}
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		stream, err := OperatorApiCmd.ShowOperator(ctx, &OperatorIn)
		if err != nil {
			fmt.Println("ShowOperator failed: ", err)
			return
		}
		objs := make([]*edgeproto.Operator, 0)
		for {
			obj, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("ShowOperator recv failed: ", err)
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
			fmt.Fprintln(output, strings.Join(OperatorHeaderSlicer(), "\t"))
			for _, obj := range objs {
				fmt.Fprintln(output, strings.Join(OperatorSlicer(obj), "\t"))
			}
			output.Flush()
		}
	},
}

func init() {
	OperatorFlagSet.StringVar(&OperatorIn.Key.Name, "key-name", "", "Key.Name")
	CreateOperatorCmd.Flags().AddFlagSet(OperatorFlagSet)
	DeleteOperatorCmd.Flags().AddFlagSet(OperatorFlagSet)
	UpdateOperatorCmd.Flags().AddFlagSet(OperatorFlagSet)
	ShowOperatorCmd.Flags().AddFlagSet(OperatorFlagSet)
}

func OperatorSetFields() {
	OperatorIn.Fields = make([]string, 0)
	if OperatorFlagSet.Lookup("key-name").Changed {
		OperatorIn.Fields = append(OperatorIn.Fields, "2.1")
	}
}
