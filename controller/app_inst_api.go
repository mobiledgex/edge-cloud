package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
)

type AppInstApi struct {
	proto.ObjStore
	appInsts    map[proto.AppInstKey]*proto.AppInst
	mux         util.Mutex
	appApi      *AppApi
	cloudletApi *CloudletApi
}

func InitAppInstApi(objStore proto.ObjStore, appApi *AppApi, cloudletApi *CloudletApi) *AppInstApi {
	api := &AppInstApi{
		ObjStore:    objStore,
		appApi:      appApi,
		cloudletApi: cloudletApi,
	}
	api.appInsts = make(map[proto.AppInstKey]*proto.AppInst)

	api.mux.Lock()
	defer api.mux.Unlock()

	err := proto.LoadAllAppInsts(api, func(obj *proto.AppInst) error {
		api.appInsts[obj.Key] = obj
		return nil
	})
	if err != nil && err == context.DeadlineExceeded {
		util.WarnLog("Init app insts failed", "error", err)
	}
	return api
}

func (s *AppInstApi) ValidateKey(key *proto.AppInstKey) error {
	if key == nil {
		return errors.New("AppInst key not specified")
	}
	if err := s.appApi.ValidateKey(&key.AppKey); err != nil {
		return err
	}
	if err := s.cloudletApi.ValidateKey(&key.CloudletKey); err != nil {
		return err
	}
	if key.Id == 0 {
		return errors.New("AppInst Id cannot be zero")
	}
	return nil
}

func (s *AppInstApi) Validate(in *proto.AppInst) error {
	if err := s.ValidateKey(&in.Key); err != nil {
		return err
	}
	if in.Liveness == proto.AppInst_UNKNOWN {
		return errors.New("Unknown liveness specified")
	}
	if !util.ValidIp(in.Ip) {
		return errors.New("Invalid IP specified")
	}
	return nil
}

func (s *AppInstApi) GetObjStoreKeyString(key *proto.AppInstKey) string {
	return GetObjStoreKey(AppInstType, key.GetKeyString())
}

func (s *AppInstApi) GetLoadKeyString() string {
	return GetObjStoreKey(AppInstType, "")
}

func (s *AppInstApi) Refresh(in *proto.AppInst, key string) error {
	s.mux.Lock()
	obj, err := proto.LoadOneAppInst(s, key)
	if err == nil {
		s.appInsts[in.Key] = obj
	} else if err == proto.ObjStoreErrKeyNotFound {
		delete(s.appInsts, in.Key)
		err = nil
	}
	defer s.mux.Unlock()
	notify.UpdateAppInst(&in.Key)
	return err
}

func (s *AppInstApi) GetAllKeys(keys map[proto.AppInstKey]bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	for key, _ := range s.appInsts {
		keys[key] = true
	}
}

func (s *AppInstApi) GetAppInst(key *proto.AppInstKey, val *proto.AppInst) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	inst, found := s.appInsts[*key]
	if found {
		*val = *inst
	}
	return found
}

func (s *AppInstApi) CreateAppInst(ctx context.Context, in *proto.AppInst) (*proto.Result, error) {
	// cache location of cloudlet in app inst
	var cloudlet proto.Cloudlet
	if s.cloudletApi.GetCloudlet(&in.Key.CloudletKey, &cloudlet) {
		in.CloudletLoc = cloudlet.Location
	}
	return in.Create(s)
}

func (s *AppInstApi) UpdateAppInst(ctx context.Context, in *proto.AppInst) (*proto.Result, error) {
	return in.Update(s)
}

func (s *AppInstApi) DeleteAppInst(ctx context.Context, in *proto.AppInst) (*proto.Result, error) {
	return in.Delete(s)
}

func (s *AppInstApi) ShowAppInst(in *proto.AppInst, cb proto.AppInstApi_ShowAppInstServer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, obj := range s.appInsts {
		if !obj.Matches(in) {
			continue
		}
		err := cb.Send(obj)
		if err != nil {
			return err
		}
	}
	return nil
}
