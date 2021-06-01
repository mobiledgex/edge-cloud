package main

import (
	"context"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

/*type AllRateLimitSettings struct {
	// RateLimitSettings for Controller Create API endpoints
	ControllerCreateApiEndpointRateLimitSettings *ratelimit.ApiEndpointRateLimitSettings
	// RateLimitSettings for Controller Show API endpoints
	ControllerShowApiEndpointRateLimitSettings *ratelimit.ApiEndpointRateLimitSettings
	// RateLimitSettings for Controller Delete API endpoints
	ControllerDeleteApiEndpointRateLimitSettings *ratelimit.ApiEndpointRateLimitSettings
	// RateLimitSettings for Controller Update API endpoints
	ControllerUpdateApiEndpointRateLimitSettings *ratelimit.ApiEndpointRateLimitSettings
	// RateLimitSettings for Controller Default API endpoints
	ControllerDefaultApiEndpointRateLimitSettings *ratelimit.ApiEndpointRateLimitSettings
	// RateLimitSettings for DME API endpoints
	DmeDefaultApiEndpointRateLimitSettings *ratelimit.ApiEndpointRateLimitSettings
}*/

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

func (r *RateLimitSettingsApi) setDefaults(ctx context.Context) error {
	defaults := edgeproto.GetDefaultRateLimitSettings()
	for _, defaultSettings := range defaults {
		_, err := r.store.Create(ctx, defaultSettings, r.sync.syncWait)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RateLimitSettingsApi) Get(key edgeproto.RateLimitSettingsKey) *edgeproto.RateLimitSettings {
	cacheData, ok := r.cache.Objs[key]
	if !ok {
		return nil
	}
	return cacheData.Obj
}

// Update RateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) UpdateRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "UpdateRateLimitSettings", "ratelimitsettings", in)
	var err error

	if err = in.Validate(edgeproto.RateLimitSettingsAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}

	if err = in.ValidateUpdateFields(); err != nil {
		return &edgeproto.Result{}, err
	}

	key := in.Key
	if err = key.ValidateKey(); err != nil {
		return nil, err
	}

	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		log.SpanLog(ctx, log.DebugLevelApi, "UpdateRateLimitSettings begin ApplySTMWait", "ratelimitsettingskey", in.Key.String())
		cur := edgeproto.RateLimitSettings{}
		if !r.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		cur.CopyInFields(in)
		r.store.STMPut(stm, &cur)

		// Update RateLimitMgrs
		services.unaryRateLimitMgr.UpdateRateLimitSettings(&cur)
		services.streamRateLimitMgr.UpdateRateLimitSettings(&cur)
		return nil
	})
	log.SpanLog(ctx, log.DebugLevelApi, "UpdateRateLimitSettings done", "ratelimitsettingskey", in.Key.String())
	return &edgeproto.Result{}, err
}

// Delete RateLimit settings for an API endpoint type (ie. no rate limiting)
func (r *RateLimitSettingsApi) DeleteRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	if !r.cache.HasKey(in.GetKey()) {
		return &edgeproto.Result{}, in.GetKey().NotFoundError()
	}
	res, err := r.store.Delete(ctx, in, r.sync.syncWait)
	if err != nil {
		return res, err
	}
	// Update RateLimitMgrs
	services.unaryRateLimitMgr.RemoveRateLimitSettings(in.Key)
	services.streamRateLimitMgr.RemoveRateLimitSettings(in.Key)
	return res, err
}

// Reset RateLimit settings to default for an API endpoint type
func (r *RateLimitSettingsApi) ResetRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	defaults := edgeproto.GetDefaultRateLimitSettings()
	for _, defaultSettings := range defaults {
		res, err := r.UpdateRateLimitSettings(ctx, defaultSettings)
		if err != nil {
			return res, err
		}
	}
	return &edgeproto.Result{}, nil
}

// Show RateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) ShowRateLimitSettings(in *edgeproto.RateLimitSettings, cb edgeproto.RateLimitSettingsApi_ShowRateLimitSettingsServer) error {
	err := r.cache.Show(in, func(obj *edgeproto.RateLimitSettings) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
