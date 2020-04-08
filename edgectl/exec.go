package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cli"
	edgecli "github.com/mobiledgex/edge-cloud/edgectl/cli"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	webrtc "github.com/pion/webrtc/v2"
	"google.golang.org/grpc"
)

var execApiCmd edgeproto.ExecApiClient

type execFunc func(context.Context, *edgeproto.ExecRequest, ...grpc.CallOption) (*edgeproto.ExecRequest, error)

func runRunCommand(c *cli.Command, args []string) error {
	return runExecRequest(c, args, execApiCmd.RunCommand)
}

func runRunConsole(c *cli.Command, args []string) error {
	return runExecRequest(c, args, execApiCmd.RunConsole)
}

func runShowLogs(c *cli.Command, args []string) error {
	return runExecRequest(c, args, execApiCmd.ShowLogs)
}

func runExecRequest(c *cli.Command, args []string, apiFunc execFunc) error {
	if execApiCmd == nil {
		return fmt.Errorf("ExecApi client not initialized")
	}
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	req := c.ReqData.(*edgeproto.ExecRequest)

	exchangeFunc := func(offer *webrtc.SessionDescription) (*edgeproto.ExecRequest, *webrtc.SessionDescription, error) {
		if offer != nil {
			offerBytes, err := json.Marshal(offer)
			if err != nil {
				return nil, nil, err
			}
			req.Offer = string(offerBytes)
		}

		ctx := context.Background()
		reply, err := apiFunc(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		if reply.Err != "" {
			return nil, nil, fmt.Errorf("%s", reply.Err)
		}
		if offer != nil {
			if reply.Answer == "" {
				return nil, nil, fmt.Errorf("empty answer")
			}
			answer := webrtc.SessionDescription{}
			err = json.Unmarshal([]byte(reply.Answer), &answer)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to unmarshal answer %s, %v",
					reply.Answer, err)
			}
			return reply, &answer, nil
		}
		return reply, nil, nil
	}
	if req.Webrtc {
		return edgecli.RunWebrtc(req, exchangeFunc, nil, edgecli.SetupLocalConsoleTunnel)
	}
	return edgecli.RunEdgeTurn(req, exchangeFunc, nil)
}
