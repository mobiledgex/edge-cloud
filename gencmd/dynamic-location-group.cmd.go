// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dynamic-location-group.proto

package gencmd

import distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import "strings"
import "time"
import "strconv"
import "github.com/spf13/cobra"
import "context"
import "os"
import "text/tabwriter"
import "github.com/spf13/pflag"
import "errors"
import "encoding/json"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/cmdsup"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var DynamicLocGroupApiCmd distributed_match_engine.DynamicLocGroupApiClient
var DlgMessageIn distributed_match_engine.DlgMessage
var DlgMessageFlagSet = pflag.NewFlagSet("DlgMessage", pflag.ExitOnError)
var DlgMessageInAckType string

func DlgMessageSlicer(in *distributed_match_engine.DlgMessage) []string {
	s := make([]string, 0, 6)
	s = append(s, strconv.FormatUint(uint64(in.Ver), 10))
	s = append(s, strconv.FormatUint(uint64(in.LgId), 10))
	s = append(s, in.GroupCookie)
	s = append(s, strconv.FormatUint(uint64(in.MessageId), 10))
	s = append(s, distributed_match_engine.DlgMessage_DlgAck_name[int32(in.AckType)])
	s = append(s, in.Message)
	return s
}

func DlgMessageHeaderSlicer() []string {
	s := make([]string, 0, 6)
	s = append(s, "Ver")
	s = append(s, "LgId")
	s = append(s, "GroupCookie")
	s = append(s, "MessageId")
	s = append(s, "AckType")
	s = append(s, "Message")
	return s
}

func DlgReplySlicer(in *distributed_match_engine.DlgReply) []string {
	s := make([]string, 0, 3)
	s = append(s, strconv.FormatUint(uint64(in.Ver), 10))
	s = append(s, strconv.FormatUint(uint64(in.AckId), 10))
	s = append(s, in.GroupCookie)
	return s
}

func DlgReplyHeaderSlicer() []string {
	s := make([]string, 0, 3)
	s = append(s, "Ver")
	s = append(s, "AckId")
	s = append(s, "GroupCookie")
	return s
}

var SendToGroupCmd = &cobra.Command{
	Use: "SendToGroup",
	Run: func(cmd *cobra.Command, args []string) {
		if DynamicLocGroupApiCmd == nil {
			fmt.Println("DynamicLocGroupApi client not initialized")
			return
		}
		var err error
		err = parseDlgMessageEnums()
		if err != nil {
			fmt.Println("SendToGroup: ", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		objs, err := DynamicLocGroupApiCmd.SendToGroup(ctx, &DlgMessageIn)
		cancel()
		if err != nil {
			fmt.Println("SendToGroup failed: ", err)
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
			fmt.Fprintln(output, strings.Join(DlgReplyHeaderSlicer(), "\t"))
			fmt.Fprintln(output, strings.Join(DlgReplySlicer(objs), "\t"))
			output.Flush()
		}
	},
}

func init() {
	DlgMessageFlagSet.Uint32Var(&DlgMessageIn.Ver, "ver", 0, "Ver")
	DlgMessageFlagSet.Uint64Var(&DlgMessageIn.LgId, "lgid", 0, "LgId")
	DlgMessageFlagSet.StringVar(&DlgMessageIn.GroupCookie, "groupcookie", "", "GroupCookie")
	DlgMessageFlagSet.Uint64Var(&DlgMessageIn.MessageId, "messageid", 0, "MessageId")
	DlgMessageFlagSet.StringVar(&DlgMessageInAckType, "acktype", "", "DlgMessageInAckType")
	DlgMessageFlagSet.StringVar(&DlgMessageIn.Message, "message", "", "Message")
	SendToGroupCmd.Flags().AddFlagSet(DlgMessageFlagSet)
}

func parseDlgMessageEnums() error {
	if DlgMessageInAckType != "" {
		switch DlgMessageInAckType {
		case "dlgackeachmessage":
			DlgMessageIn.AckType = distributed_match_engine.DlgMessage_DlgAck(0)
		case "dlgasyeverynmessage":
			DlgMessageIn.AckType = distributed_match_engine.DlgMessage_DlgAck(1)
		case "dlgnoack":
			DlgMessageIn.AckType = distributed_match_engine.DlgMessage_DlgAck(2)
		default:
			return errors.New("Invalid value for DlgMessageInAckType")
		}
	}
	return nil
}
