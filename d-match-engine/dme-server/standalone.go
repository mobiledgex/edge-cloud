package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type standaloneServer struct{}

func (s *standaloneServer) CreateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	addApp(in)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) DeleteAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	removeApp(in)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) UpdateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	addApp(in)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	appInst := edgeproto.AppInst{}

	tbl := carrierAppTbl
	tbl.Lock()
	defer tbl.Unlock()

	for _, a := range tbl.apps {
		for _, c := range a.insts {
			appInst.Key.AppKey = a.key.appKey
			appInst.Key.CloudletKey = c.cloudletKey
			appInst.Uri = c.uri
			appInst.CloudletLoc = c.location
			err := cb.Send(&appInst)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
