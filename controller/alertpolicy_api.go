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
type AlertPolicyApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.AlertPolicyStore
	cache edgeproto.AlertPolicyCache
}

func NewAlertPolicyApi(sync *Sync, all *AllApis) *AlertPolicyApi {
	alertPolicyApi := AlertPolicyApi{}
	alertPolicyApi.all = all
	alertPolicyApi.sync = sync
	alertPolicyApi.store = edgeproto.NewAlertPolicyStore(sync.store)
	edgeproto.InitAlertPolicyCache(&alertPolicyApi.cache)
	sync.RegisterCache(&alertPolicyApi.cache)
	return &alertPolicyApi
}

func (a *AlertPolicyApi) CreateAlertPolicy(ctx context.Context, in *edgeproto.AlertPolicy) (*edgeproto.Result, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "CreateAlertPolicy", "alert", in.Key.String())
	var err error

	if err = in.Validate(edgeproto.AlertPolicyAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}

	// Protect against user defined alerts that can oscillate too quickly
	if in.TriggerTime < a.all.settingsApi.Get().AlertPolicyMinTriggerTime {
		return &edgeproto.Result{},
			fmt.Errorf("Trigger time cannot be less than %s",
				a.all.settingsApi.Get().AlertPolicyMinTriggerTime.TimeDuration().String())
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
	log.SpanLog(ctx, log.DebugLevelApi, "CreateAlertPolicy done", "alert", in.Key.String())
	return &edgeproto.Result{}, err
}

func (a *AlertPolicyApi) DeleteAlertPolicy(ctx context.Context, in *edgeproto.AlertPolicy) (res *edgeproto.Result, reterr error) {
	err := a.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.AlertPolicy{}
		if !a.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if cur.DeletePrepare {
			return fmt.Errorf("AlertPolicy already being deleted")
		}
		cur.DeletePrepare = true
		a.store.STMPut(stm, &cur)
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}
	defer func() {
		if reterr == nil {
			return
		}
		err := a.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			cur := edgeproto.AlertPolicy{}
			if !a.store.STMGet(stm, &in.Key, &cur) {
				return nil
			}
			if cur.DeletePrepare {
				cur.DeletePrepare = false
				a.store.STMPut(stm, &cur)
			}
			return nil
		})
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "undo delete prepare failed", "key", in.Key, "err", err)
		}
	}()

	if appKey := a.all.appApi.UsesAlertPolicy(&in.Key); appKey != nil {
		return &edgeproto.Result{}, fmt.Errorf("Alert is in use by App %s", appKey.GetKeyString())
	}
	return a.store.Delete(ctx, in, a.sync.syncWait)
}

func (a *AlertPolicyApi) ShowAlertPolicy(in *edgeproto.AlertPolicy, cb edgeproto.AlertPolicyApi_ShowAlertPolicyServer) error {
	err := a.cache.Show(in, func(obj *edgeproto.AlertPolicy) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (a *AlertPolicyApi) UpdateAlertPolicy(ctx context.Context, in *edgeproto.AlertPolicy) (*edgeproto.Result, error) {
	cur := edgeproto.AlertPolicy{}
	changed := 0
	err := a.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !a.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed = cur.CopyInFields(in)
		if changed == 0 {
			return nil
		}
		if err := cur.Validate(nil); err != nil {
			return err
		}
		// Protect against user defined alerts that can oscillate too quickly
		if cur.TriggerTime < a.all.settingsApi.Get().AlertPolicyMinTriggerTime {
			return fmt.Errorf("Trigger time cannot be less than %s",
				a.all.settingsApi.Get().AlertPolicyMinTriggerTime.TimeDuration().String())
		}

		if !cloudcommon.IsAlertSeverityValid(cur.Severity) {
			return fmt.Errorf("Invalid severity. Valid severities: %s", cloudcommon.GetValidAlertSeverityString())
		}
		a.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}
