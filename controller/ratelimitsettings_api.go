package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon/ratelimit"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type RateLimitSettingsApi struct {
	sync         *Sync
	flowstore    edgeproto.FlowRateLimitSettingsStore
	flowcache    edgeproto.FlowRateLimitSettingsCache
	maxreqsstore edgeproto.MaxReqsRateLimitSettingsStore
	maxreqscache edgeproto.MaxReqsRateLimitSettingsCache
}

var rateLimitSettingsApi = RateLimitSettingsApi{}

func InitRateLimitSettingsApi(sync *Sync) {
	rateLimitSettingsApi.sync = sync
	// Init store and cache for FlowRateLimitSettings
	rateLimitSettingsApi.flowstore = edgeproto.NewFlowRateLimitSettingsStore(sync.store)
	edgeproto.InitFlowRateLimitSettingsCache(&rateLimitSettingsApi.flowcache)
	sync.RegisterCache(&rateLimitSettingsApi.flowcache)
	// Init store and cache for MaxReqsRateLimitSettings
	rateLimitSettingsApi.maxreqsstore = edgeproto.NewMaxReqsRateLimitSettingsStore(sync.store)
	edgeproto.InitMaxReqsRateLimitSettingsCache(&rateLimitSettingsApi.maxreqscache)
	sync.RegisterCache(&rateLimitSettingsApi.maxreqscache)
}

// Store initial default Flow and MaxReqs RateLimitSettings
func (r *RateLimitSettingsApi) initDefaultRateLimitSettings(ctx context.Context) error {
	err := r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		defaultsettings := edgeproto.GetDefaultRateLimitSettings()
		for _, dsetting := range defaultsettings {
			// Store default FlowRateLimitSettings
			for name, fsetting := range dsetting.FlowSettings {
				flowRateLimitSettings := buildFlowRateLimitSettings(dsetting.Key, name, fsetting)
				buf := edgeproto.FlowRateLimitSettings{}
				if !r.flowstore.STMGet(stm, &flowRateLimitSettings.Key, &buf) {
					r.flowstore.STMPut(stm, flowRateLimitSettings)
				}
			}

			// Store default MaxReqsRateLimitSettings
			for name, msetting := range dsetting.MaxReqsSettings {
				maxReqsRateLimitSettings := buildMaxReqsRateLimitSettings(dsetting.Key, name, msetting)
				buf := edgeproto.MaxReqsRateLimitSettings{}
				if !r.maxreqsstore.STMGet(stm, &maxReqsRateLimitSettings.Key, &buf) {
					r.maxreqsstore.STMPut(stm, maxReqsRateLimitSettings)
				}
			}
		}
		return nil
	})
	return err
}

func buildFlowRateLimitSettings(key edgeproto.RateLimitSettingsKey, name string, f *edgeproto.FlowSettings) *edgeproto.FlowRateLimitSettings {
	return &edgeproto.FlowRateLimitSettings{
		Key: edgeproto.FlowRateLimitSettingsKey{
			FlowSettingsName: name,
			RateLimitKey:     key,
		},
		Settings: *f,
	}
}

func buildMaxReqsRateLimitSettings(key edgeproto.RateLimitSettingsKey, name string, m *edgeproto.MaxReqsSettings) *edgeproto.MaxReqsRateLimitSettings {
	return &edgeproto.MaxReqsRateLimitSettings{
		Key: edgeproto.MaxReqsRateLimitSettingsKey{
			MaxReqsSettingsName: name,
			RateLimitKey:        key,
		},
		Settings: *m,
	}
}

// Show RateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) ShowRateLimitSettings(in *edgeproto.RateLimitSettings, cb edgeproto.RateLimitSettingsApi_ShowRateLimitSettingsServer) error {
	if settingsApi.Get().DisableRateLimit {
		return fmt.Errorf("DisableRateLimit must be false to ShowRateLimitSettings")
	}
	// Get all FlowRateLimitSettings with corresponding RateLimitKey
	flowsettings := make([]*edgeproto.FlowRateLimitSettings, 0)
	ffilter := &edgeproto.FlowRateLimitSettings{
		Key: edgeproto.FlowRateLimitSettingsKey{
			RateLimitKey: in.Key,
		},
	}
	err := r.flowcache.Show(ffilter, func(obj *edgeproto.FlowRateLimitSettings) error {
		flowsettings = append(flowsettings, obj)
		return nil
	})
	if err != nil {
		return err
	}

	// Get all MaxReqsRateLimitSettings with corresponding RateLimitKey
	maxreqssettings := make([]*edgeproto.MaxReqsRateLimitSettings, 0)
	mfilter := &edgeproto.MaxReqsRateLimitSettings{
		Key: edgeproto.MaxReqsRateLimitSettingsKey{
			RateLimitKey: in.Key,
		},
	}
	err = r.maxreqscache.Show(mfilter, func(obj *edgeproto.MaxReqsRateLimitSettings) error {
		maxreqssettings = append(maxreqssettings, obj)
		return nil
	})
	if err != nil {
		return err
	}

	ratelimitsettings := ratelimit.DbToUserSettings(flowsettings, maxreqssettings)
	for _, settings := range ratelimitsettings {
		if err = cb.Send(settings); err != nil {
			return err
		}
	}

	return nil
}

// Create FlowRateLimitSettings for the specified RateLimitSettings. If no RateLimitSettings exists, create a new one
func (r *RateLimitSettingsApi) CreateFlowRateLimitSettings(ctx context.Context, in *edgeproto.FlowRateLimitSettings) (*edgeproto.Result, error) {
	if settingsApi.Get().DisableRateLimit {
		return nil, fmt.Errorf("DisableRateLimit must be false to CreateFlowRateLimitSettings")
	}

	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "CreateFlowRateLimitSettings", "flowratelimitsettings", in)

	cur := edgeproto.FlowRateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if r.flowstore.STMGet(stm, &in.Key, &cur) {
			return in.Key.ExistsError()
		}

		// Validate fields and key before storing
		if err = in.Validate(edgeproto.FlowRateLimitSettingsAllFieldsMap); err != nil {
			return err
		}

		log.SpanLog(ctx, log.DebugLevelApi, "Add new FlowRateLimitSettings to RateLimitSettings", "key", in.Key.String())

		r.flowstore.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

// Update FlowRateLimitSettings for the specified RateLimitSettings
func (r *RateLimitSettingsApi) UpdateFlowRateLimitSettings(ctx context.Context, in *edgeproto.FlowRateLimitSettings) (*edgeproto.Result, error) {
	if settingsApi.Get().DisableRateLimit {
		return nil, fmt.Errorf("DisableRateLimit must be false to UpdateFlowRateLimitSettings")
	}

	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "UpdateFlowRateLimitSettings", "flowratelimitsettings", in)

	cur := edgeproto.FlowRateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !r.flowstore.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}

		changes := cur.CopyInFields(in)
		if changes == 0 {
			return nil
		}

		// Validate fields before storing
		if err = cur.Validate(edgeproto.FlowRateLimitSettingsAllFieldsMap); err != nil {
			return err
		}

		log.SpanLog(ctx, log.DebugLevelApi, "Updating FlowRateLimitSettings for RateLimitSettings", "key", in.Key.String(), "updated settings", cur)
		r.flowstore.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

// Delete FlowRateLimitSettings for the specified RateLimitSettings. If no FlowSettings and MaxReqsSettings left, remove the RateLimitSettings
func (r *RateLimitSettingsApi) DeleteFlowRateLimitSettings(ctx context.Context, in *edgeproto.FlowRateLimitSettings) (*edgeproto.Result, error) {
	if settingsApi.Get().DisableRateLimit {
		return nil, fmt.Errorf("DisableRateLimit must be false to DeleteFlowRateLimitSettings")
	}

	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "DeleteFlowRateLimitSettings", "key", in.Key.String())

	buf := &edgeproto.FlowRateLimitSettings{}
	if !r.flowcache.Get(&in.Key, buf) {
		return nil, in.Key.NotFoundError()
	}
	return r.flowstore.Delete(ctx, in, r.sync.syncWait)
}

// Show FlowRateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) ShowFlowRateLimitSettings(in *edgeproto.FlowRateLimitSettings, cb edgeproto.RateLimitSettingsApi_ShowFlowRateLimitSettingsServer) error {
	if settingsApi.Get().DisableRateLimit {
		if *testMode {
			return nil
		}
		return fmt.Errorf("DisableRateLimit must be false to ShowFlowRateLimitSettings")
	}

	err := r.flowcache.Show(in, func(obj *edgeproto.FlowRateLimitSettings) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

// Create MaxReqsRateLimitSettings for the specified RateLimitSettings. If no RateLimitSettings exists, create a new one
func (r *RateLimitSettingsApi) CreateMaxReqsRateLimitSettings(ctx context.Context, in *edgeproto.MaxReqsRateLimitSettings) (*edgeproto.Result, error) {
	if settingsApi.Get().DisableRateLimit {
		return nil, fmt.Errorf("DisableRateLimit must be false to CreateMaxReqsRateLimitSettings")
	}

	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "CreateMaxReqsRateLimitSettings", "maxreqsratelimitsettings", in)

	cur := edgeproto.MaxReqsRateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if r.maxreqsstore.STMGet(stm, &in.Key, &cur) {
			return in.Key.ExistsError()
		}

		// Validate fields and key before storing
		if err = in.Validate(edgeproto.MaxReqsRateLimitSettingsAllFieldsMap); err != nil {
			return err
		}

		log.SpanLog(ctx, log.DebugLevelApi, "Add new MaxReqsRateLimitSettings to RateLimitSettings", "key", in.Key.String())

		r.maxreqsstore.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

// Update MaxReqsRateLimitSettings for the specified RateLimitSettings
func (r *RateLimitSettingsApi) UpdateMaxReqsRateLimitSettings(ctx context.Context, in *edgeproto.MaxReqsRateLimitSettings) (*edgeproto.Result, error) {
	if settingsApi.Get().DisableRateLimit {
		return nil, fmt.Errorf("DisableRateLimit must be false to UpdateMaxReqsRateLimitSettings")
	}

	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "UpdateMaxReqsRateLimitSettings", "maxreqsratelimitsettings", in)

	cur := edgeproto.MaxReqsRateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !r.maxreqsstore.STMGet(stm, &in.Key, &cur) {
			return in.Key.RateLimitKey.NotFoundError()
		}

		changes := cur.CopyInFields(in)
		if changes == 0 {
			return nil
		}

		// Validate fields before storing
		if err = cur.Validate(edgeproto.MaxReqsRateLimitSettingsAllFieldsMap); err != nil {
			return err
		}

		log.SpanLog(ctx, log.DebugLevelApi, "Updating MaxReqsRateLimitSettings for RateLimitSettings", "key", in.Key.String(), "updated settings", cur)
		r.maxreqsstore.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

// Delete MaxReqsRateLimitSettings for the specified RateLimitSettings. If no FlowSettings and MaxReqsSettings left, remove the RateLimitSettings
func (r *RateLimitSettingsApi) DeleteMaxReqsRateLimitSettings(ctx context.Context, in *edgeproto.MaxReqsRateLimitSettings) (*edgeproto.Result, error) {
	if settingsApi.Get().DisableRateLimit {
		return nil, fmt.Errorf("DisableRateLimit must be false to DeleteMaxReqsRateLimitSettings")
	}

	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "DeleteMaxReqsRateLimitSettings", "key", in.Key.String())

	buf := &edgeproto.MaxReqsRateLimitSettings{}
	if !r.maxreqscache.Get(&in.Key, buf) {
		return nil, in.Key.NotFoundError()
	}
	return r.maxreqsstore.Delete(ctx, in, r.sync.syncWait)
}

// Show MaxReqsRateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) ShowMaxReqsRateLimitSettings(in *edgeproto.MaxReqsRateLimitSettings, cb edgeproto.RateLimitSettingsApi_ShowMaxReqsRateLimitSettingsServer) error {
	if settingsApi.Get().DisableRateLimit {
		if *testMode {
			return nil
		}
		return fmt.Errorf("DisableRateLimit must be false to ShowMaxReqsRateLimitSettings")
	}

	err := r.maxreqscache.Show(in, func(obj *edgeproto.MaxReqsRateLimitSettings) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
