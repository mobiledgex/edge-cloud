package main

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
	rateLimitSettingsApi.store = edgeproto.NewRateLimitettingsStore(sync.store)
	edgeproto.InitRateLimitSettingsCache(&rateLimitSettingsApi.cache)
	sync.RegisterCache(&rateLimitSettingsApi.cache)
}
