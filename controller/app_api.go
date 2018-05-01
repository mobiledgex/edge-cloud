// app config

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
}

func InitAppApi(objStore proto.ObjStore) *AppApi {
	api := &AppApi{ObjStore: objStore}
	api.apps = make(map[proto.AppKey]*proto.App)

	api.mux.Lock()
	defer api.mux.Unlock()

	key := proto.App{}
	err := objStore.List(api.GetKeyString(&key), func(key, val []byte) error {
		var app proto.App
		err := json.Unmarshal(val, &app)
		if err != nil {
			util.WarnLog("Failed to parse app data", "val", string(val))
			return nil
		}
		api.apps[*app.Key] = &app
		return nil
	})
	if err != nil {
		util.WarnLog("Init apps failed", "error", err)
	}
	return api
}

func (s *AppApi) ValidateKey(in *proto.App) error {
	if in.Key == nil {
		errors.New("App key not specified")
	}
	if !util.ValidName(in.Key.DevName) {
		return errors.New("Invalid app developer name")
	}
	if !util.ValidName(in.Key.AppName) {
		return errors.New("Invalid app name")
	}
	if !util.ValidName(in.Key.Version) {
		return errors.New("Invalid app version string")
	}
	return nil
}

func (s *AppApi) Validate(in *proto.App) error {
	return s.ValidateKey(in)
}

func (s *AppApi) GetKeyString(in *proto.App) string {
	var str string
	key := in.Key
	if key == nil || key.DevName == "" {
		str = ""
	} else if key.AppName == "" {
		str = key.DevName
	} else if key.Version == "" {
		str = fmt.Sprintf("%s/%s", key.DevName, key.AppName)
	} else {
		str = fmt.Sprintf("%s/%s/%s", key.DevName, key.AppName, key.Version)
	}
	return GetObjStoreKey(AppType, str)
}

func (s *AppApi) Refresh(in *proto.App, key string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	val, _, err := s.ObjStore.Get(key)
	if err == nil {
		var app proto.App
		err = json.Unmarshal(val, &app)
		if err != nil {
			util.DebugLog(util.DebugLevelApi, "Failed to parse app data", "val", string(val))
			return err
		}
		s.apps[*in.Key] = &app
	} else if err == proto.ObjStoreErrKeyNotFound {
		delete(s.apps, *in.Key)
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
