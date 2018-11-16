package main

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"golang.org/x/net/context"
)

type standaloneServer struct{}

func (s *standaloneServer) CreateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	addApp(in)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) DeleteApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	removeApp(in)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) UpdateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	addApp(in)
	return &edgeproto.Result{}, nil
}

func (s *standaloneServer) ShowApp(in *edgeproto.App, cb edgeproto.AppApi_ShowAppServer) error {
	app := edgeproto.App{}

	tbl := dmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()

	for key, a := range tbl.apps {
		a.Lock()
		app.Key = key
		app.AuthPublicKey = a.authPublicKey
		app.AndroidPackageName = a.androidPackageName
		a.Unlock()
		err := cb.Send(&app)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *standaloneServer) CreateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) error {
	addAppInst(in)
	return nil
}

func (s *standaloneServer) DeleteAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	removeAppInst(in)
	return nil
}

func (s *standaloneServer) UpdateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_UpdateAppInstServer) error {
	addAppInst(in)
	return nil
}

func (s *standaloneServer) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	appInst := edgeproto.AppInst{}

	tbl := dmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()

	for _, a := range tbl.apps {
		for _, c := range a.carriers {
			for _, i := range c.insts {
				appInst.Key.AppKey = a.appKey
				appInst.Key.CloudletKey = i.cloudletKey
				appInst.Uri = i.uri
				appInst.CloudletLoc = i.location
				err := cb.Send(&appInst)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
