package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cli"
	webrtcshell "github.com/mobiledgex/edge-cloud/edgectl/cli"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	webrtc "github.com/pion/webrtc/v2"
)

var execApiCmd edgeproto.ExecApiClient

func runExecRequest(c *cli.Command, args []string) error {
	if execApiCmd == nil {
		return fmt.Errorf("ExecApi client not initialized")
	}
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	req := c.ReqData.(*edgeproto.ExecRequest)

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
	return webrtcshell.RunWebrtcShell(exchangeFunc)
}
