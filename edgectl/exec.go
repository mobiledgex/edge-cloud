// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"

	"github.com/edgexr/edge-cloud/cli"
	edgecli "github.com/edgexr/edge-cloud/edgectl/cli"
	"github.com/edgexr/edge-cloud/edgeproto"
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
	options := &edgecli.ExecOptions{
		Stdin: cli.Interactive,
		Tty:   cli.Tty,
	}
	return edgecli.RunEdgeTurn(req, options, exchangeFunc)
}
