// useralert config
package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
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

	// Protect against user defined alerts that can oscillate too quickly
	if in.TriggerTime < settingsApi.Get().UserDefinedAlertMinTriggerTime {
		return &edgeproto.Result{},
			fmt.Errorf("Trigger time cannot be less than %s",
				settingsApi.Get().UserDefinedAlertMinTriggerTime.TimeDuration().String())
	}

	if !cloudcommon.IsAlertSeverityValid(in.Severity) {
		return &edgeproto.Result{},
			fmt.Errorf("Invalid severity. Valid severities: %s", cloudcommon.GetValidAlertSeverityString())
	}

	err = a.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if a.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		a.store.STMPut(stm, in)
		return nil
	})
	log.SpanLog(ctx, log.DebugLevelApi, "CreateUserAlert done", "alert", in.Key.String())
	return &edgeproto.Result{}, err
}

func (a *UserAlertApi) DeleteUserAlert(ctx context.Context, in *edgeproto.UserAlert) (*edgeproto.Result, error) {
	if !a.cache.HasKey(&in.Key) {
		return &edgeproto.Result{}, in.Key.NotFoundError()
	}

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
	cur := edgeproto.UserAlert{}
	changed := 0
	err := a.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !a.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed = cur.CopyInFields(in)
		if changed == 0 {
			return nil
		}
		if err := cur.Validate(edgeproto.UserAlertAllFieldsMap); err != nil {
			return err
		}
		// Protect against user defined alerts that can oscillate too quickly
		if in.TriggerTime < settingsApi.Get().UserDefinedAlertMinTriggerTime {
			return fmt.Errorf("Trigger time cannot be less than %s",
				settingsApi.Get().UserDefinedAlertMinTriggerTime.TimeDuration().String())
		}

		if !cloudcommon.IsAlertSeverityValid(cur.Severity) {
			return fmt.Errorf("Invalid severity. Valid severities: %s", cloudcommon.GetValidAlertSeverityString())
		}
		a.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}
