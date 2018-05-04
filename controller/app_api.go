// app config

package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
)

// Should only be one of these instantiated in main
type AppApi struct {
	// object database - methods of interface are promoted to this object
	proto.ObjStore
	// list of all apps
	apps map[proto.AppKey]*proto.App
	// table lock
	mux util.Mutex
	// reference to developers
	devApi *DeveloperApi
}

func InitAppApi(objStore proto.ObjStore, devApi *DeveloperApi) *AppApi {
	api := &AppApi{ObjStore: objStore}
	api.apps = make(map[proto.AppKey]*proto.App)
	api.devApi = devApi

	api.mux.Lock()
	defer api.mux.Unlock()

	err := proto.LoadAllApps(api, func(obj *proto.App) error {
		api.apps[obj.Key] = obj
		return nil
	})
	if err != nil {
		util.WarnLog("Init apps failed", "error", err)
	}
	return api
}

func (s *AppApi) ValidateKey(key *proto.AppKey) error {
	if key == nil {
		return errors.New("App key not specified")
	}
	if err := s.devApi.ValidateKey(&key.DeveloperKey); err != nil {
		return err
	}
	if !s.devApi.HasDeveloper(&key.DeveloperKey) {
		return errors.New("Specified developer not found")
	}
	if !util.ValidName(key.Name) {
		return errors.New("Invalid app name")
	}
	if !util.ValidName(key.Version) {
		return errors.New("Invalid app version string")
	}
	return nil
}

func (s *AppApi) Validate(in *proto.App) error {
	// TODO: validate other fields?
	return s.ValidateKey(&in.Key)
}

func (s *AppApi) HasApp(key *proto.AppKey) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, found := s.apps[*key]
	return found
}

func (s *AppApi) GetObjStoreKeyString(key *proto.AppKey) string {
	return GetObjStoreKey(AppType, key.GetKeyString())
}

func (s *AppApi) GetLoadKeyString() string {
	return GetObjStoreKey(AppType, "")
}

func (s *AppApi) Refresh(in *proto.App, key string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	obj, err := proto.LoadOneApp(s, key)
	if err == nil {
		s.apps[in.Key] = obj
	} else if err == proto.ObjStoreErrKeyNotFound {
		delete(s.apps, in.Key)
		err = nil
	}
	return err
}

func (s *AppApi) CreateApp(ctx context.Context, in *proto.App) (*proto.Result, error) {
	return in.Create(s)
}

func (s *AppApi) UpdateApp(ctx context.Context, in *proto.App) (*proto.Result, error) {
	return in.Update(s)
}

func (s *AppApi) DeleteApp(ctx context.Context, in *proto.App) (*proto.Result, error) {
	return in.Delete(s)
}

func (s *AppApi) ShowApp(in *proto.App, cb proto.AppApi_ShowAppServer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, myApp := range s.apps {
		if !myApp.Matches(in) {
			continue
		}
		err := cb.Send(myApp)
		if err != nil {
			return err
		}
	}
	return nil
}
