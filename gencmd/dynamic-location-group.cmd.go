// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dynamic-location-group.proto

package gencmd

import distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import "strings"
import "strconv"
import "github.com/spf13/cobra"
import "context"
import "os"
import "text/tabwriter"
import "github.com/spf13/pflag"
import "errors"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/cmdsup"
import "google.golang.org/grpc/status"
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
var DlgMessageNoConfigFlagSet = pflag.NewFlagSet("DlgMessageNoConfig", pflag.ExitOnError)
var DlgMessageInAckType string
var DlgAckStrings = []string{
	"DlgAckEachMessage",
	"DlgAsyEveryNMessage",
	"DlgNoAck",
}

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

func DlgMessageWriteOutputArray(objs []*distributed_match_engine.DlgMessage) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(DlgMessageHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(DlgMessageSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func DlgMessageWriteOutputOne(obj *distributed_match_engine.DlgMessage) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(DlgMessageHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(DlgMessageSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
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

func DlgReplyWriteOutputArray(objs []*distributed_match_engine.DlgReply) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(DlgReplyHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(DlgReplySlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func DlgReplyWriteOutputOne(obj *distributed_match_engine.DlgReply) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(DlgReplyHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(DlgReplySlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}

var SendToGroupCmd = &cobra.Command{
	Use: "SendToGroup",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseDlgMessageEnums()
		if err != nil {
			return fmt.Errorf("SendToGroup failed: %s", err.Error())
		}
		return SendToGroup(&DlgMessageIn)
	},
}

func SendToGroup(in *distributed_match_engine.DlgMessage) error {
	if DynamicLocGroupApiCmd == nil {
		return fmt.Errorf("DynamicLocGroupApi client not initialized")
	}
	ctx := context.Background()
	obj, err := DynamicLocGroupApiCmd.SendToGroup(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("SendToGroup failed: %s", errstr)
	}
	DlgReplyWriteOutputOne(obj)
	return nil
}

func SendToGroups(data []distributed_match_engine.DlgMessage, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("SendToGroup %v\n", data[ii])
		myerr := SendToGroup(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DynamicLocGroupApiCmds = []*cobra.Command{
	SendToGroupCmd,
}

func init() {
	DlgMessageFlagSet.Uint32Var(&DlgMessageIn.Ver, "ver", 0, "Ver")
	DlgMessageFlagSet.Uint64Var(&DlgMessageIn.LgId, "lgid", 0, "LgId")
	DlgMessageFlagSet.StringVar(&DlgMessageIn.GroupCookie, "groupcookie", "", "GroupCookie")
	DlgMessageFlagSet.Uint64Var(&DlgMessageIn.MessageId, "messageid", 0, "MessageId")
	DlgMessageFlagSet.StringVar(&DlgMessageInAckType, "acktype", "", "one of [DlgAckEachMessage DlgAsyEveryNMessage DlgNoAck]")
	DlgMessageFlagSet.StringVar(&DlgMessageIn.Message, "message", "", "Message")
	SendToGroupCmd.Flags().AddFlagSet(DlgMessageFlagSet)
}

func DynamicLocGroupApiAllowNoConfig() {
	SendToGroupCmd.Flags().AddFlagSet(DlgMessageNoConfigFlagSet)
}

func parseDlgMessageEnums() error {
	if DlgMessageInAckType != "" {
		switch DlgMessageInAckType {
		case "DlgAckEachMessage":
			DlgMessageIn.AckType = distributed_match_engine.DlgMessage_DlgAck(0)
		case "DlgAsyEveryNMessage":
			DlgMessageIn.AckType = distributed_match_engine.DlgMessage_DlgAck(1)
		case "DlgNoAck":
			DlgMessageIn.AckType = distributed_match_engine.DlgMessage_DlgAck(2)
		default:
			return errors.New("Invalid value for DlgMessageInAckType")
		}
	}
	return nil
}
