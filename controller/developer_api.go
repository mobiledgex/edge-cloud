package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
)

type DeveloperApi struct {
	// database - methods of interface are promoted to this object
	proto.ObjStore
	// list of all developers
	developers map[proto.DeveloperKey]*proto.Developer
	// table lock
	mux util.Mutex
}

func InitDeveloperApi(objStore proto.ObjStore) *DeveloperApi {
	api := &DeveloperApi{ObjStore: objStore}
	api.developers = make(map[proto.DeveloperKey]*proto.Developer)

	api.mux.Lock()
	defer api.mux.Unlock()

	// read existing data into memory
	key := proto.Developer{}
	err := objStore.List(api.GetKeyString(&key), func(key, val []byte) error {
		var dev proto.Developer
		err := json.Unmarshal(val, &dev)
		if err != nil {
			util.WarnLog("Failed to parse developer data", "val", string(val))
			return nil
		}
		api.developers[*dev.Key] = &dev
		return nil
	})
	if err != nil {
		util.WarnLog("Init developers failed", "error", err)
	}
	return api
}

func (s *DeveloperApi) ValidateKey(in *proto.Developer) error {
	if in.Key == nil {
		errors.New("Developer key not specified")
	}
	if !util.ValidName(in.Key.Name) {
		errors.New("invalid developer name")
	}
	return nil
}

func (s *DeveloperApi) Validate(in *proto.Developer) error {
	return s.ValidateKey(in)
}

func (s *DeveloperApi) GetKeyString(in *proto.Developer) string {
	if in.Key == nil {
		return GetObjStoreKey(DeveloperType, "")
	}
	return GetObjStoreKey(DeveloperType, in.Key.Name)
}

func (s *DeveloperApi) Refresh(in *proto.Developer, key string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	val, _, err := s.ObjStore.Get(key)
	if err == nil {
		var dev proto.Developer
		err = json.Unmarshal(val, &dev)
		if err != nil {
			util.DebugLog(util.DebugLevelApi, "Failed to parse developer data", "val", string(val))
			return err
		}
		s.developers[*in.Key] = &dev
	} else if err == proto.ObjStoreErrKeyNotFound {
		delete(s.developers, *in.Key)
		err = nil
	}
	return err
}

func (s *DeveloperApi) CreateDeveloper(ctx context.Context, in *proto.Developer) (*proto.Result, error) {
	return in.Create(s)
}

func (s *DeveloperApi) UpdateDeveloper(ctx context.Context, in *proto.Developer) (*proto.Result, error) {
	return in.Update(s)
}

func (s *DeveloperApi) DeleteDeveloper(ctx context.Context, in *proto.Developer) (*proto.Result, error) {
	return in.Delete(s)
}

func (s *DeveloperApi) ShowDeveloper(in *proto.Developer, cb proto.DeveloperApi_ShowDeveloperServer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	// keys may be empty to allow showing multiple developers
	for _, myDev := range s.developers {
		if !myDev.Matches(in) {
			continue
		}
		err := cb.Send(myDev)
		if err != nil {
			return err
		}
	}
	return nil
}
