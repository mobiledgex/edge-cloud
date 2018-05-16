package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
)

type CloudletApi struct {
	edgeproto.ObjStore
	cloudlets   map[edgeproto.CloudletKey]*edgeproto.Cloudlet
	mux         util.Mutex
	operatorApi *OperatorApi
}

func InitCloudletApi(objStore edgeproto.ObjStore, opApi *OperatorApi) *CloudletApi {
	api := &CloudletApi{
		ObjStore:    objStore,
		operatorApi: opApi,
	}
	api.cloudlets = make(map[edgeproto.CloudletKey]*edgeproto.Cloudlet)

	api.mux.Lock()
	defer api.mux.Unlock()

	err := edgeproto.LoadAllCloudlets(api, func(obj *edgeproto.Cloudlet) error {
		api.cloudlets[obj.Key] = obj
		return nil
	})
	if err != nil && err == context.DeadlineExceeded {
		util.WarnLog("Init cloudlets failed", "error", err)
	}
	return api
}

func (s *CloudletApi) ValidateKey(key *edgeproto.CloudletKey) error {
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

func (s *CloudletApi) Validate(in *edgeproto.Cloudlet) error {
	if err := s.ValidateKey(&in.Key); err != nil {
		return err
	}
	if in.AccessIp != nil && !util.ValidIp(in.AccessIp) {
		return errors.New("Invalid access ip format")
	}
	return nil
}

func (s *CloudletApi) GetObjStoreKeyString(key *edgeproto.CloudletKey) string {
	return GetObjStoreKey(CloudletType, key.GetKeyString())
}

func (s *CloudletApi) GetLoadKeyString() string {
	return GetObjStoreKey(CloudletType, "")
}

func (s *CloudletApi) Refresh(in *edgeproto.Cloudlet, key string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	obj, err := edgeproto.LoadOneCloudlet(s, key)
	if err == nil {
		s.cloudlets[in.Key] = obj
	} else if err == edgeproto.ObjStoreErrKeyNotFound {
		delete(s.cloudlets, in.Key)
		err = nil
	}
	notify.UpdateCloudlet(&in.Key)
	// TODO: If location changed, update location in all associate app insts
	return err
}

func (s *CloudletApi) GetAllKeys(keys map[edgeproto.CloudletKey]bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	for key, _ := range s.cloudlets {
		keys[key] = true
	}
}

func (s *CloudletApi) GetCloudlet(key *edgeproto.CloudletKey, val *edgeproto.Cloudlet) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	inst, found := s.cloudlets[*key]
	if found {
		*val = *inst
	}
	return found
}

func (s *CloudletApi) CreateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	return in.Create(s)
}

func (s *CloudletApi) UpdateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	return in.Update(s)
}

func (s *CloudletApi) DeleteCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	return in.Delete(s)
}

func (s *CloudletApi) ShowCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_ShowCloudletServer) error {
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
