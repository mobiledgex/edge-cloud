// useralert config
package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// Should only be one of these instantiated in main
type UserAlertApi struct {
	sync  *Sync
	store edgeproto.UserAlertStore
	cache edgeproto.UserAlertCache
}

var userAlertApi = UserAlertApi{}

func InitUserAlertApi(sync *Sync) {
	userAlertApi.sync = sync
	userAlertApi.store = edgeproto.NewUserAlertStore(sync.store)
	edgeproto.InitUserAlertCache(&userAlertApi.cache)
	sync.RegisterCache(&userAlertApi.cache)
}

func (a *UserAlertApi) CreateUserAlert(ctx context.Context, in *edgeproto.UserAlert) (*edgeproto.Result, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "CreateUserAlert", "alert", in.Key.String())
	var err error

	if err = in.Validate(edgeproto.UserAlertAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}
	// Since active connections and other metrics are part
	// of different instances of Prometheus, disallow mixing them
	if in.ActiveConnLimit != 0 {
		if in.CpuLimit != 0 || in.MemLimit != 0 || in.DiskLimit != 0 {
			return &edgeproto.Result{},
				fmt.Errorf("Active Connection Alerts should not include any other triggers.")
		}
	}
	err = a.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		log.SpanLog(ctx, log.DebugLevelApi, "CreateUserAlert begin ApplySTMWait", "alert", in.Key.String())
		if a.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		a.store.STMPut(stm, in)
		return nil
	})
	log.SpanLog(ctx, log.DebugLevelApi, "CreateAppCreateUserAlert done", "alert", in.Key.String())
	return &edgeproto.Result{}, err
}

func (a *UserAlertApi) DeleteUserAlert(ctx context.Context, in *edgeproto.UserAlert) (*edgeproto.Result, error) {
	if appApi.UsesUserDefinedAlert(&in.Key) {
		return &edgeproto.Result{}, fmt.Errorf("Alert is in use by App")
	}
	return a.store.Delete(ctx, in, a.sync.syncWait)
}

func (a *UserAlertApi) ShowUserAlert(in *edgeproto.UserAlert, cb edgeproto.UserAlertApi_ShowUserAlertServer) error {
	err := a.cache.Show(in, func(obj *edgeproto.UserAlert) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (a *UserAlertApi) UpdateUserAlert(ctx context.Context, in *edgeproto.UserAlert) (*edgeproto.Result, error) {
	// TODO
	return &edgeproto.Result{}, nil
}
