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
	return nil, nil
}

// Update RateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) UpdateRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	if !settingsApi.Get().RateLimitEnable {
		return nil, fmt.Errorf("RateLimitEnable must be true in order to UpdateRateLimitSettings")
	}
	log.SpanLog(ctx, log.DebugLevelApi, "UpdateRateLimitSettings", "ratelimitsettings", in)
	var err error

	// Validate key
	key := in.Key
	if err = key.ValidateKey(); err != nil {
		return nil, err
	}

	err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.RateLimitSettings{}
		if !r.store.STMGet(stm, &in.Key, &cur) {
			log.SpanLog(ctx, log.DebugLevelApi, "Adding new RateLimitSettings", "key", in.Key.String(), "settings", in)
			cur = *in
		} else {
			cur.CopyInFields(in)
			log.SpanLog(ctx, log.DebugLevelApi, "Updating previous RateLimitSettings", "key", in.Key.String(), "updated settings", cur)
		}
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
	if !settingsApi.Get().RateLimitEnable {
		return nil, fmt.Errorf("RateLimitEnable must be true in order to DeleteRateLimitSettings")
	}
	log.SpanLog(ctx, log.DebugLevelApi, "DeleteRateLimitSettings", "key", in.Key.String())
	var err error

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
	return res, err
}

// Reset RateLimit settings to default for an API endpoint type
func (r *RateLimitSettingsApi) ResetRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	if !settingsApi.Get().RateLimitEnable {
		return nil, fmt.Errorf("RateLimitEnable must be true in order to ResetRateLimitSettings")
	}
	log.SpanLog(ctx, log.DebugLevelApi, "ResetRateLimitSettings", "key", in.Key.String())
	var err error

	// Validate key
	key := in.Key
	if err := key.ValidateKey(); err != nil {
		return nil, err
	}

	// Get the Default RateLimitSettings associated with the key
	var newSettings *edgeproto.RateLimitSettings
	dmedefaults := edgeproto.GetDefaultDmeRateLimitSettings()
	newSettings, _ = dmedefaults[in.Key]

	// Reset to default settings if found
	if newSettings != nil {
		err = r.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			r.store.STMPut(stm, newSettings)
			return nil
		})
	}
	return &edgeproto.Result{}, err
}

// Show RateLimit settings for an API endpoint type
func (r *RateLimitSettingsApi) ShowRateLimitSettings(in *edgeproto.RateLimitSettings, cb edgeproto.RateLimitSettingsApi_ShowRateLimitSettingsServer) error {
	if !settingsApi.Get().RateLimitEnable {
		return fmt.Errorf("RateLimitEnable must be true in order to ShowRateLimitSettings")
	}
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
