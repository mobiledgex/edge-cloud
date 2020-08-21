package main

import (
	"context"
	"fmt"
	"io"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstLatencyApi struct {
	sync *Sync
}

var appInstLatencyApi = AppInstLatencyApi{}

func InitAppInstLatencyApi(sync *Sync) {
	appInstLatencyApi.sync = sync
}

func (s *AppInstLatencyApi) RequestAppInstLatency(ctx context.Context, in *edgeproto.AppInstLatency) (*edgeproto.Result, error) {
	conn, err := notifyRootConnect(ctx, *notifyParentAddrs)
	if err != nil {
		return nil, err
	}
	client := edgeproto.NewAppInstLatencyApiClient(conn)
	ctx, cancel := context.WithTimeout(ctx, node.DefaultDebugTimeout)
	defer cancel()

	return client.RequestAppInstLatency(ctx, in)
}

func (s *AppInstLatencyApi) ShowAppInstLatency(in *edgeproto.AppInstLatency, cb edgeproto.AppInstLatencyApi_ShowAppInstLatencyServer) error {
	conn, err := notifyRootConnect(cb.Context(), *notifyParentAddrs)
	if err != nil {
		return err
	}
	client := edgeproto.NewAppInstLatencyApiClient(conn)
	ctx, cancel := context.WithTimeout(cb.Context(), node.DefaultDebugTimeout)
	defer cancel()

	stream, err := client.ShowAppInstLatency(ctx, in)
	if err != nil {
		return err
	}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ShowappInstLatency failed, %v", err)
		}
		err = cb.Send(obj)
		if err != nil {
			return err
		}
	}
	return nil
}
