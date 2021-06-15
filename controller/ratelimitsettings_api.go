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

// Gets the RateLimitSettings that corresponds to the specified RateLimitSettingsKey
func (r *RateLimitSettingsApi) Get(key edgeproto.RateLimitSettingsKey) *edgeproto.RateLimitSettings {
	cacheData, ok := r.cache.Objs[key]
	if !ok {
		return nil
	}
	return cacheData.Obj
}

func (r *RateLimitSettingsApi) CreateRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	err := validateRateLimitSettings(in)
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "UpdateRateLimitSettings", "ratelimitsettings", in)

	cur := edgeproto.RateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if r.store.STMGet(stm, &in.Key, &cur) {
			return fmt.Errorf("Unable to CreateRateLimitSettings - key %v already exists", in.Key)
		}

		// Set cur to the incoming RateLimitSettings struct
		cur = *in
		log.SpanLog(ctx, log.DebugLevelApi, "Adding new RateLimitSettings", "key", in.Key.String(), "settings", in)

		// Validate fields and key before storing
		if err = cur.Validate(edgeproto.RateLimitSettingsAllFieldsMap); err != nil {
			return err
		}
		r.store.STMPut(stm, &cur)
		return nil
	})
	if err == nil {
		// Add RateLimitSettings in RateLimitMgr
		services.rateLimitManager.UpdateRateLimitSettings(&cur)
	}
	return &edgeproto.Result{}, err
}

// Update RateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) UpdateRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	err := validateRateLimitSettings(in)
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "UpdateRateLimitSettings", "ratelimitsettings", in)

	cur := edgeproto.RateLimitSettings{}
	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !r.store.STMGet(stm, &in.Key, &cur) {
			return fmt.Errorf("Unable to UpdateRateLimitSettings - key %v not found", in.Key)
		}

		cur.CopyInFields(in)
		log.SpanLog(ctx, log.DebugLevelApi, "Updating previous RateLimitSettings", "key", in.Key.String(), "updated settings", cur)

		// Validate fields and key before storing
		if err = cur.Validate(edgeproto.RateLimitSettingsAllFieldsMap); err != nil {
			return err
		}
		r.store.STMPut(stm, &cur)
		return nil
	})
	if err == nil {
		// Update RateLimitSettings in RateLimitMgr
		services.rateLimitManager.UpdateRateLimitSettings(&cur)
	}
	return &edgeproto.Result{}, err
}

// Delete RateLimit settings for an API endpoint type (ie. no rate limiting)
func (r *RateLimitSettingsApi) DeleteRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	err := validateRateLimitSettings(in)
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "DeleteRateLimitSettings", "key", in.Key.String())

	res, err := r.store.Delete(ctx, in, r.sync.syncWait)
	if err != nil {
		return res, err
	}
	// Remove RateLimitSettings from RateLimitMgr
	services.rateLimitManager.RemoveRateLimitSettings(in.Key)
	return res, err
}

// Show RateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) ShowRateLimitSettings(in *edgeproto.RateLimitSettings, cb edgeproto.RateLimitSettingsApi_ShowRateLimitSettingsServer) error {
	err := r.cache.Show(in, func(obj *edgeproto.RateLimitSettings) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

// Helper function that validates the incoming RateLimitSettings request with current settings
func validateRateLimitSettings(in *edgeproto.RateLimitSettings) error {
	// Check that DisableDmeRateLimit is false if ApiEndpointType is Dme
	if settingsApi.Get().DisableDmeRateLimit && in.Key.ApiEndpointType == edgeproto.ApiEndpointType_DME {
		return fmt.Errorf("DisableDmeRateLimit in settings must be false for ApiEndpointType Dme")
	}

	// Check that DisableCtrlRateLimit is false if ApiEndpointType is Controller
	if settingsApi.Get().DisableCtrlRateLimit && in.Key.ApiEndpointType == edgeproto.ApiEndpointType_CONTROLLER {
		return fmt.Errorf("DisableCtrlRateLimit in settings must be false for ApiEndpointType Controller")
	}

	if in.Key.RateLimitTarget == edgeproto.RateLimitTarget_PER_USER {
		return fmt.Errorf("PerUser rate limiting is not implemented for %v apis", in.Key.ApiEndpointType)
	}

	// Validate key
	key := in.Key
	return key.ValidateKey()
}

// Notify callbacks

// This should only be called when DME updates its RateLimitSettings based on
func (r *RateLimitSettingsApi) Update(ctx context.Context, in *edgeproto.RateLimitSettings, rev int64) {
	r.CreateRateLimitSettings(ctx, in)
}

func (r *RateLimitSettingsApi) Delete(ctx context.Context, in *edgeproto.RateLimitSettings, rev int64) {
}

func (r *RateLimitSettingsApi) Prune(ctx context.Context, keys map[edgeproto.RateLimitSettingsKey]struct{}) {
}

func (r *RateLimitSettingsApi) Flush(ctx context.Context, notifyId int64) {
}
