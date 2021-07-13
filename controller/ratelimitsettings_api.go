package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type RateLimitSettingsApi struct {
	sync  *Sync
	store edgeproto.RateLimitSettingsStore
	cache edgeproto.RateLimitSettingsCache
}

var rateLimitSettingsApi = RateLimitSettingsApi{}

func InitRateLimitSettingsApi(sync *Sync) {
	rateLimitSettingsApi.sync = sync
	rateLimitSettingsApi.store = edgeproto.NewRateLimitSettingsStore(sync.store)
	edgeproto.InitRateLimitSettingsCache(&rateLimitSettingsApi.cache)
	sync.RegisterCache(&rateLimitSettingsApi.cache)
}

// Store initial default RateLimitSettings
func (r *RateLimitSettingsApi) initDefaultRateLimitSettings(ctx context.Context) error {
	err := r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		defaultsettings := edgeproto.GetDefaultRateLimitSettings()
		for _, settings := range defaultsettings {
			buf := edgeproto.RateLimitSettings{}
			if !r.store.STMGet(stm, &settings.Key, &buf) {
				r.store.STMPut(stm, settings)
			}
		}
		return nil
	})
	return err
}

// Gets the RateLimitSettings that corresponds to the specified RateLimitSettingsKey
func (r *RateLimitSettingsApi) Get(key edgeproto.RateLimitSettingsKey) *edgeproto.RateLimitSettings {
	buf := &edgeproto.RateLimitSettings{}
	if !r.cache.Get(&key, buf) {
		return nil
	}
	return buf
}

// Create RateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) CreateRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	err := validateRateLimitSettings(in)
	if err != nil {
		return nil, err
	}

	if (in.FlowSettings == nil || len(in.FlowSettings) == 0) && (in.MaxReqsSettings == nil || len(in.MaxReqsSettings) == 0) {
		return nil, fmt.Errorf("One of FlowSettings or MaxReqsSettings must be set in order to create RateLimitSettings")
	}

	log.SpanLog(ctx, log.DebugLevelApi, "CreateRateLimitSettings", "ratelimitsettings", in)

	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.RateLimitSettings{}
		if r.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.ExistsError()
		}

		log.SpanLog(ctx, log.DebugLevelApi, "Adding new RateLimitSettings", "key", in.Key.String(), "settings", in)

		// Validate fields before storing
		if err = in.Validate(nil); err != nil {
			return err
		}
		r.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

// Delete RateLimit settings for an API endpoint type (ie. no rate limiting)
func (r *RateLimitSettingsApi) DeleteRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "DeleteRateLimitSettings", "key", in.Key.String())

	buf := &edgeproto.RateLimitSettings{}
	if !r.cache.Get(&in.Key, buf) {
		return nil, in.Key.NotFoundError()
	}
	return r.store.Delete(ctx, in, r.sync.syncWait)
}

// Show RateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) ShowRateLimitSettings(in *edgeproto.RateLimitSettings, cb edgeproto.RateLimitSettingsApi_ShowRateLimitSettingsServer) error {
	err := r.cache.Show(in, func(obj *edgeproto.RateLimitSettings) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

// Create FlowRateLimitSettings for the specified RateLimitSettings. If no RateLimitSettings exists, create a new one
func (r *RateLimitSettingsApi) CreateFlowRateLimitSettings(ctx context.Context, in *edgeproto.FlowRateLimitSettings) (*edgeproto.Result, error) {
	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "CreateFlowRateLimitSettings", "flowratelimitsettings", in)

	cur := edgeproto.RateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !r.store.STMGet(stm, &in.Key.RateLimitKey, &cur) {
			log.SpanLog(ctx, log.DebugLevelApi, "Cannot find RateLimitSettings, creating a new entry")
			cur.Key = in.Key.RateLimitKey
		}

		if cur.FlowSettings == nil {
			cur.FlowSettings = make(map[string]*edgeproto.FlowSettings)
		}

		_, ok := cur.FlowSettings[in.Key.FlowSettingsName]
		if ok {
			return in.Key.ExistsError()
		}

		// Validate fields and key before storing
		if err = in.Validate(edgeproto.FlowRateLimitSettingsAllFieldsMap); err != nil {
			return err
		}

		cur.FlowSettings[in.Key.FlowSettingsName] = in.Settings
		log.SpanLog(ctx, log.DebugLevelApi, "Add new FlowRateLimitSettings to RateLimitSettings", "key", in.Key.String(), "updated settings", cur)

		r.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

// Update FlowRateLimitSettings for the specified RateLimitSettings
func (r *RateLimitSettingsApi) UpdateFlowRateLimitSettings(ctx context.Context, in *edgeproto.FlowRateLimitSettings) (*edgeproto.Result, error) {
	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "UpdateFlowRateLimitSettings", "flowratelimitsettings", in)

	cur := edgeproto.RateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !r.store.STMGet(stm, &in.Key.RateLimitKey, &cur) {
			return in.Key.RateLimitKey.NotFoundError()
		}

		fsettings, ok := cur.FlowSettings[in.Key.FlowSettingsName]
		if !ok {
			return in.Key.NotFoundError()
		}

		changes := fsettings.CopyInFields(in.Settings)
		if changes == 0 {
			return nil
		}
		// Validate fields before storing
		if err = fsettings.Validate(); err != nil {
			return err
		}

		log.SpanLog(ctx, log.DebugLevelApi, "Updating FlowRateLimitSettings for RateLimitSettings", "key", in.Key.String(), "updated settings", cur)
		r.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

// Delete FlowRateLimitSettings for the specified RateLimitSettings. If no FlowSettings and MaxReqsSettings left, remove the RateLimitSettings
func (r *RateLimitSettingsApi) DeleteFlowRateLimitSettings(ctx context.Context, in *edgeproto.FlowRateLimitSettings) (*edgeproto.Result, error) {
	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "DeleteFlowRateLimitSettings", "flowratelimitsettings", in)

	cur := edgeproto.RateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !r.store.STMGet(stm, &in.Key.RateLimitKey, &cur) {
			return in.Key.RateLimitKey.NotFoundError()
		}

		_, ok := cur.FlowSettings[in.Key.FlowSettingsName]
		if !ok {
			return in.Key.NotFoundError()
		}

		delete(cur.FlowSettings, in.Key.FlowSettingsName)
		log.SpanLog(ctx, log.DebugLevelApi, "Deleting FlowRateLimitSettings from RateLimitSettings", "key", in.Key.String(), "updated settings", cur)

		// remove entry if no FlowSettings and no MaxReqsSettings
		if (cur.FlowSettings == nil || len(cur.FlowSettings) == 0) && (cur.MaxReqsSettings == nil || len(cur.MaxReqsSettings) == 0) {
			r.store.STMDel(stm, &cur.Key)
		} else {
			r.store.STMPut(stm, &cur)
		}
		return nil
	})
	return &edgeproto.Result{}, err
}

// Create MaxReqsRateLimitSettings for the specified RateLimitSettings. If no RateLimitSettings exists, create a new one
func (r *RateLimitSettingsApi) CreateMaxReqsRateLimitSettings(ctx context.Context, in *edgeproto.MaxReqsRateLimitSettings) (*edgeproto.Result, error) {
	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "CreateMaxReqsRateLimitSettings", "maxreqsratelimitsettings", in)

	cur := edgeproto.RateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !r.store.STMGet(stm, &in.Key.RateLimitKey, &cur) {
			log.SpanLog(ctx, log.DebugLevelApi, "Cannot find RateLimitSettings, creating a new entry")
			cur.Key = in.Key.RateLimitKey
		}

		if cur.MaxReqsSettings == nil {
			cur.MaxReqsSettings = make(map[string]*edgeproto.MaxReqsSettings)
		}

		_, ok := cur.MaxReqsSettings[in.Key.MaxReqsSettingsName]
		if ok {
			return in.Key.ExistsError()
		}

		// Validate fields and key before storing
		if err = in.Validate(edgeproto.MaxReqsRateLimitSettingsAllFieldsMap); err != nil {
			return err
		}

		cur.MaxReqsSettings[in.Key.MaxReqsSettingsName] = in.Settings
		log.SpanLog(ctx, log.DebugLevelApi, "Add new MaxReqsRateLimitSettings to RateLimitSettings", "key", in.Key.String(), "updated settings", cur)

		r.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

// Update MaxReqsRateLimitSettings for the specified RateLimitSettings
func (r *RateLimitSettingsApi) UpdateMaxReqsRateLimitSettings(ctx context.Context, in *edgeproto.MaxReqsRateLimitSettings) (*edgeproto.Result, error) {
	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "UpdateMaxReqsRateLimitSettings", "maxreqsratelimitsettings", in)

	cur := edgeproto.RateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !r.store.STMGet(stm, &in.Key.RateLimitKey, &cur) {
			return in.Key.RateLimitKey.NotFoundError()
		}

		msettings, ok := cur.MaxReqsSettings[in.Key.MaxReqsSettingsName]
		if !ok {
			return in.Key.NotFoundError()
		}

		changes := msettings.CopyInFields(in.Settings)
		if changes == 0 {
			return nil
		}
		// Validate fields before storing
		if err = msettings.Validate(); err != nil {
			return err
		}

		log.SpanLog(ctx, log.DebugLevelApi, "Updating MaxReqsRateLimitSettings for RateLimitSettings", "key", in.Key.String(), "updated settings", cur)
		r.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

// Delete MaxReqsRateLimitSettings for the specified RateLimitSettings. If no FlowSettings and MaxReqsSettings left, remove the RateLimitSettings
func (r *RateLimitSettingsApi) DeleteMaxReqsRateLimitSettings(ctx context.Context, in *edgeproto.MaxReqsRateLimitSettings) (*edgeproto.Result, error) {
	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "DeleteMaxReqsRateLimitSettings", "maxreqsratelimitsettings", in)

	cur := edgeproto.RateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !r.store.STMGet(stm, &in.Key.RateLimitKey, &cur) {
			return in.Key.RateLimitKey.NotFoundError()
		}

		_, ok := cur.MaxReqsSettings[in.Key.MaxReqsSettingsName]
		if !ok {
			return in.Key.NotFoundError()
		}

		delete(cur.MaxReqsSettings, in.Key.MaxReqsSettingsName)
		log.SpanLog(ctx, log.DebugLevelApi, "Deleting MaxReqsRateLimitSettings from RateLimitSettings", "key", in.Key.String(), "updated settings", cur)

		// remove entry if no FlowSettings and no MaxReqsSettings
		if (cur.FlowSettings == nil || len(cur.FlowSettings) == 0) && (cur.MaxReqsSettings == nil || len(cur.MaxReqsSettings) == 0) {
			r.store.STMDel(stm, &cur.Key)
		} else {
			r.store.STMPut(stm, &cur)
		}
		return nil
	})
	return &edgeproto.Result{}, err
}

// Helper function that validates the incoming RateLimitSettings request with current settings
func validateRateLimitSettings(in *edgeproto.RateLimitSettings) error {
	// Check that DisableDmeRateLimit is false if ApiEndpointType is Dme
	if settingsApi.Get().DisableDmeRateLimit && in.Key.ApiEndpointType == edgeproto.ApiEndpointType_DME {
		return fmt.Errorf("DisableDmeRateLimit in settings must be false for ApiEndpointType Dme")
	}

	if in.Key.RateLimitTarget == edgeproto.RateLimitTarget_PER_USER {
		return fmt.Errorf("PerUser rate limiting is not implemented for %v apis", in.Key.ApiEndpointType)
	}

	// Validate key
	return in.Key.ValidateKey()
}
