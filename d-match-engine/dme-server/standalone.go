package main

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type standaloneServer struct{}

func (s *standaloneServer) CreateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) error {
	addApp(in)
	return nil
}

func (s *standaloneServer) DeleteAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	removeApp(in)
	return nil
}

func (s *standaloneServer) UpdateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_UpdateAppInstServer) error {
	addApp(in)
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
