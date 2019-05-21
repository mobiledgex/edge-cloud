package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgectl/cli"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/gencmd"
	webrtc "github.com/pion/webrtc/v2"
	"github.com/spf13/cobra"
)

var execApiCmd edgeproto.ExecApiClient

func NewExecRequestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "RunCommand",
		Short: "Run a command or shell on an AppInst",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExecRequest(&gencmd.ExecRequestIn)
		},
	}
	cmd.Flags().AddFlagSet(gencmd.ExecRequestFlagSet)
	return cmd
}

func runExecRequest(req *edgeproto.ExecRequest) error {
	if execApiCmd == nil {
		return fmt.Errorf("ExecApi client not initialized")
	}
	exchangeFunc := func(offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
		offerBytes, err := json.Marshal(&offer)
		if err != nil {
			return nil, err
		}
		req.Offer = string(offerBytes)

		ctx := context.Background()
		reply, err := execApiCmd.RunCommand(ctx, req)
		if err != nil {
			return nil, err
		}
		if reply.Err != "" {
			return nil, fmt.Errorf("%s", reply.Err)
		}
		if reply.Answer == "" {
			return nil, fmt.Errorf("empty answer")
		}
		answer := webrtc.SessionDescription{}
		err = json.Unmarshal([]byte(reply.Answer), &answer)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal answer %s, %v",
				reply.Answer, err)
		}
		return &answer, nil
	}
	return cli.RunWebrtcShell(exchangeFunc)
}
