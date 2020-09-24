package main

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cli"
	edgecli "github.com/mobiledgex/edge-cloud/edgectl/cli"
	"github.com/mobiledgex/edge-cloud/edgeproto"
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

func runAccessCloudlet(c *cli.Command, args []string) error {
	return runExecRequest(c, args, execApiCmd.AccessCloudlet)
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

	exchangeFunc := func() (*edgeproto.ExecRequest, error) {
		ctx := context.Background()
		reply, err := apiFunc(ctx, req)
		if err != nil {
			return nil, err
		}
		if reply.Err != "" {
			return nil, fmt.Errorf("%s", reply.Err)
		}
		return reply, nil
	}
	return edgecli.RunEdgeTurn(req, exchangeFunc)
}
