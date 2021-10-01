package ratelimit

import "github.com/mobiledgex/edge-cloud/edgeproto"

// Convert db-based objects to user-based objects
func DbToUserSettings(fsettings []*edgeproto.FlowRateLimitSettings, msettings []*edgeproto.MaxReqsRateLimitSettings) []*edgeproto.RateLimitSettings {
	settingsmap := make(map[edgeproto.RateLimitSettingsKey]*edgeproto.RateLimitSettings)

	for _, fsetting := range fsettings {
		key := fsetting.Key.RateLimitKey
		ratelimitsetting, ok := settingsmap[key]
		if !ok || ratelimitsetting == nil {
			ratelimitsetting = &edgeproto.RateLimitSettings{
				Key: key,
			}
			settingsmap[key] = ratelimitsetting
		}
		if ratelimitsetting.FlowSettings == nil {
			ratelimitsetting.FlowSettings = make(map[string]*edgeproto.FlowSettings)
		}
		ratelimitsetting.FlowSettings[fsetting.Key.FlowSettingsName] = &fsetting.Settings
	}

	for _, msetting := range msettings {
		key := msetting.Key.RateLimitKey
		ratelimitsetting, ok := settingsmap[key]
		if !ok || ratelimitsetting == nil {
			ratelimitsetting = &edgeproto.RateLimitSettings{
				Key: key,
			}
			settingsmap[key] = ratelimitsetting
		}
		if ratelimitsetting.MaxReqsSettings == nil {
			ratelimitsetting.MaxReqsSettings = make(map[string]*edgeproto.MaxReqsSettings)
		}
		ratelimitsetting.MaxReqsSettings[msetting.Key.MaxReqsSettingsName] = &msetting.Settings
	}

	ratelimitsettings := make([]*edgeproto.RateLimitSettings, 0)
	for _, settings := range settingsmap {
		ratelimitsettings = append(ratelimitsettings, settings)
	}
	return ratelimitsettings
}

// Convert user-based objects to db-based objects
func UserToDbSettings(settings []*edgeproto.RateLimitSettings) ([]*edgeproto.FlowRateLimitSettings, []*edgeproto.MaxReqsRateLimitSettings) {
	fsettings := []*edgeproto.FlowRateLimitSettings{}
	msettings := []*edgeproto.MaxReqsRateLimitSettings{}

	for _, s := range settings {
		for name, flowSettings := range s.FlowSettings {
			frset := edgeproto.FlowRateLimitSettings{
				Key: edgeproto.FlowRateLimitSettingsKey{
					FlowSettingsName: name,
					RateLimitKey:     s.Key,
				},
				Settings: *flowSettings,
			}
			fsettings = append(fsettings, &frset)
		}
		for name, maxSettings := range s.MaxReqsSettings {
			mrset := edgeproto.MaxReqsRateLimitSettings{
				Key: edgeproto.MaxReqsRateLimitSettingsKey{
					MaxReqsSettingsName: name,
					RateLimitKey:        s.Key,
				},
				Settings: *maxSettings,
			}
			msettings = append(msettings, &mrset)
		}
	}
	return fsettings, msettings
}
