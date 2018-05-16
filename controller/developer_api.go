package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

type DeveloperApi struct {
	// database - methods of interface are promoted to this object
	edgeproto.ObjStore
	// list of all developers
	developers map[edgeproto.DeveloperKey]*edgeproto.Developer
	// table lock
	mux util.Mutex
}

func InitDeveloperApi(objStore edgeproto.ObjStore) *DeveloperApi {
	api := &DeveloperApi{ObjStore: objStore}
	api.developers = make(map[edgeproto.DeveloperKey]*edgeproto.Developer)

	api.mux.Lock()
	defer api.mux.Unlock()

	err := edgeproto.LoadAllDevelopers(api, func(obj *edgeproto.Developer) error {
		api.developers[obj.Key] = obj
		return nil
	})
	if err != nil && err == context.DeadlineExceeded {
		util.WarnLog("Init developers failed", "error", err)
	}
	return api
}

func (s *DeveloperApi) ValidateKey(key *edgeproto.DeveloperKey) error {
	if key == nil {
		return errors.New("Developer key not specified")
	}
	if !util.ValidName(key.Name) {
		return errors.New("invalid developer name")
	}
	return nil
}

func (s *DeveloperApi) Validate(in *edgeproto.Developer) error {
	return s.ValidateKey(&in.Key)
}

func (s *DeveloperApi) HasDeveloper(key *edgeproto.DeveloperKey) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, found := s.developers[*key]
	return found
}

func (s *DeveloperApi) GetObjStoreKeyString(key *edgeproto.DeveloperKey) string {
	return GetObjStoreKey(DeveloperType, key.GetKeyString())
}

func (s *DeveloperApi) GetLoadKeyString() string {
	return GetObjStoreKey(DeveloperType, "")
}

func (s *DeveloperApi) Refresh(in *edgeproto.Developer, key string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	obj, err := edgeproto.LoadOneDeveloper(s, key)
	if err == nil {
		s.developers[in.Key] = obj
	} else if err == edgeproto.ObjStoreErrKeyNotFound {
		delete(s.developers, in.Key)
		err = nil
	}
	return err
}

func (s *DeveloperApi) CreateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return in.Create(s)
}

func (s *DeveloperApi) UpdateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return in.Update(s)
}

func (s *DeveloperApi) DeleteDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return in.Delete(s)
}

func (s *DeveloperApi) ShowDeveloper(in *edgeproto.Developer, cb edgeproto.DeveloperApi_ShowDeveloperServer) error {
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
