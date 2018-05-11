package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
)

type CloudletApi struct {
	proto.ObjStore
	cloudlets   map[proto.CloudletKey]*proto.Cloudlet
	mux         util.Mutex
	operatorApi *OperatorApi
}

func InitCloudletApi(objStore proto.ObjStore, opApi *OperatorApi) *CloudletApi {
	api := &CloudletApi{
		ObjStore:    objStore,
		operatorApi: opApi,
	}
	api.cloudlets = make(map[proto.CloudletKey]*proto.Cloudlet)

	api.mux.Lock()
	defer api.mux.Unlock()

	err := proto.LoadAllCloudlets(api, func(obj *proto.Cloudlet) error {
		api.cloudlets[obj.Key] = obj
		return nil
	})
	if err != nil && err == context.DeadlineExceeded {
		util.WarnLog("Init cloudlets failed", "error", err)
	}
	return api
}

func (s *CloudletApi) ValidateKey(key *proto.CloudletKey) error {
	if key == nil {
		return errors.New("Cloudlet key not specified")
	}
	if err := s.operatorApi.ValidateKey(&key.OperatorKey); err != nil {
		return err
	}
	if !s.operatorApi.HasOperator(&key.OperatorKey) {
		return errors.New("Specified cloudlet operator not found")
	}
	if !util.ValidName(key.Name) {
		return errors.New("Invalid cloudlet name")
	}
	return nil
}

func (s *CloudletApi) Validate(in *proto.Cloudlet) error {
	if err := s.ValidateKey(&in.Key); err != nil {
		return err
	}
	if in.AccessIp != nil && !util.ValidIp(in.AccessIp) {
		return errors.New("Invalid access ip format")
	}
	return nil
}

func (s *CloudletApi) GetObjStoreKeyString(key *proto.CloudletKey) string {
	return GetObjStoreKey(CloudletType, key.GetKeyString())
}

func (s *CloudletApi) GetLoadKeyString() string {
	return GetObjStoreKey(CloudletType, "")
}

func (s *CloudletApi) Refresh(in *proto.Cloudlet, key string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	obj, err := proto.LoadOneCloudlet(s, key)
	if err == nil {
		s.cloudlets[in.Key] = obj
	} else if err == proto.ObjStoreErrKeyNotFound {
		delete(s.cloudlets, in.Key)
		err = nil
	}
	notify.UpdateCloudlet(&in.Key)
	// TODO: If location changed, update location in all associate app insts
	return err
}

func (s *CloudletApi) GetAllKeys(keys map[proto.CloudletKey]bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	for key, _ := range s.cloudlets {
		keys[key] = true
	}
}

func (s *CloudletApi) GetCloudlet(key *proto.CloudletKey, val *proto.Cloudlet) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	inst, found := s.cloudlets[*key]
	if found {
		*val = *inst
	}
	return found
}

func (s *CloudletApi) CreateCloudlet(ctx context.Context, in *proto.Cloudlet) (*proto.Result, error) {
	return in.Create(s)
}

func (s *CloudletApi) UpdateCloudlet(ctx context.Context, in *proto.Cloudlet) (*proto.Result, error) {
	return in.Update(s)
}

func (s *CloudletApi) DeleteCloudlet(ctx context.Context, in *proto.Cloudlet) (*proto.Result, error) {
	return in.Delete(s)
}

func (s *CloudletApi) ShowCloudlet(in *proto.Cloudlet, cb proto.CloudletApi_ShowCloudletServer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, obj := range s.cloudlets {
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
