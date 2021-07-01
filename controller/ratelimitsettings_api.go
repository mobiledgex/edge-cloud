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

// Update RateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) CreateRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	err := validateRateLimitSettings(in)
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "CreateRateLimitSettings", "ratelimitsettings", in)

	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.RateLimitSettings{}
		if r.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.ExistsError()
		}

		log.SpanLog(ctx, log.DebugLevelApi, "Adding new RateLimitSettings", "key", in.Key.String(), "settings", in)

		// Validate fields and key before storing
		if err = in.Validate(edgeproto.RateLimitSettingsAllFieldsMap); err != nil {
			return err
		}
		r.store.STMPut(stm, in)
		return nil
	})
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
			return in.Key.NotFoundError()
		}

		// Copy updated fields into cur
		cur.CopyInFields(in)
		log.SpanLog(ctx, log.DebugLevelApi, "Updating previous RateLimitSettings", "key", in.Key.String(), "updated settings", cur)

		// Validate fields and key before storing
		if err = cur.Validate(edgeproto.RateLimitSettingsAllFieldsMap); err != nil {
			return err
		}
		r.store.STMPut(stm, &cur)
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
