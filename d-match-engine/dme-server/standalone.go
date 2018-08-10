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
