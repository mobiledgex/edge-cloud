package main

import (
	"context"

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

// Store Default Controller RateLimitSettings (if removeRateLimit, don't store anything)
func (r *RateLimitSettingsApi) setControllerDefaults(ctx context.Context) error {
	if !*removeRateLimit {
		err := r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			defaults := edgeproto.GetDefaultControllerRateLimitSettings()
			for _, defaultSettings := range defaults {
				r.store.STMPut(stm, defaultSettings)
			}
			return nil
		})
		return err
	}
	return nil
}

// Gets the RateLimitSettings that corresponds to the specified RateLimitSettingsKey
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

	// Validate fields and key
	if err = in.Validate(edgeproto.RateLimitSettingsAllFieldsMap); err != nil {
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
			log.SpanLog(ctx, log.DebugLevelApi, "Adding new RateLimitSettings", "key", in.Key.String(), "settings", in)
			cur = *in
		} else {
			cur.CopyInFields(in)
			log.SpanLog(ctx, log.DebugLevelApi, "Updating previous RateLimitSettings", "key", in.Key.String(), "updated settings", cur)
		}

		r.store.STMPut(stm, &cur)
		// Update RateLimitMgrs
		services.unaryRateLimitMgr.UpdateRateLimitSettings(&cur)
		services.streamRateLimitMgr.UpdateRateLimitSettings(&cur)
		return nil
	})
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
	// Update RateLimitMgrs with "empty" RateLimitSettings
	in = &edgeproto.RateLimitSettings{
		Key: in.Key,
	}
	services.unaryRateLimitMgr.UpdateRateLimitSettings(in)
	services.streamRateLimitMgr.UpdateRateLimitSettings(in)
	return res, err
}

// Reset RateLimit settings to default for an API endpoint type
func (r *RateLimitSettingsApi) ResetRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	ctrldefaults := edgeproto.GetDefaultControllerRateLimitSettings()
	for _, defaultSettings := range ctrldefaults {
		if in.Key == defaultSettings.Key {
			res, err := r.UpdateRateLimitSettings(ctx, defaultSettings)
			if err != nil {
				return res, err
			}
		}
	}
	dmedefaults := edgeproto.GetDefaultDmeRateLimitSettings()
	for _, defaultSettings := range dmedefaults {
		if in.Key == defaultSettings.Key {
			res, err := r.UpdateRateLimitSettings(ctx, defaultSettings)
			if err != nil {
				return res, err
			}
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

// Notify callbacks
func (r *RateLimitSettingsApi) Update(ctx context.Context, in *edgeproto.RateLimitSettings, rev int64) {
	r.UpdateRateLimitSettings(ctx, in)
}

func (r *RateLimitSettingsApi) Delete(ctx context.Context, in *edgeproto.RateLimitSettings, rev int64) {
}

func (r *RateLimitSettingsApi) Prune(ctx context.Context, keys map[edgeproto.RateLimitSettingsKey]struct{}) {
}

func (r *RateLimitSettingsApi) Flush(ctx context.Context, notifyId int64) {
}
