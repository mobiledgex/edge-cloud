package main

import (
	"context"
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

	err := proto.LoadAllDevelopers(api, func(obj *proto.Developer) error {
		api.developers[obj.Key] = obj
		return nil
	})
	if err != nil && err == context.DeadlineExceeded {
		util.WarnLog("Init developers failed", "error", err)
	}
	return api
}

func (s *DeveloperApi) ValidateKey(key *proto.DeveloperKey) error {
	if key == nil {
		return errors.New("Developer key not specified")
	}
	if !util.ValidName(key.Name) {
		return errors.New("invalid developer name")
	}
	return nil
}

func (s *DeveloperApi) Validate(in *proto.Developer) error {
	return s.ValidateKey(&in.Key)
}

func (s *DeveloperApi) HasDeveloper(key *proto.DeveloperKey) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, found := s.developers[*key]
	return found
}

func (s *DeveloperApi) GetObjStoreKeyString(key *proto.DeveloperKey) string {
	return GetObjStoreKey(DeveloperType, key.GetKeyString())
}

func (s *DeveloperApi) GetLoadKeyString() string {
	return GetObjStoreKey(DeveloperType, "")
}

func (s *DeveloperApi) Refresh(in *proto.Developer, key string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	obj, err := proto.LoadOneDeveloper(s, key)
	if err == nil {
		s.developers[in.Key] = obj
	} else if err == proto.ObjStoreErrKeyNotFound {
		delete(s.developers, in.Key)
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
