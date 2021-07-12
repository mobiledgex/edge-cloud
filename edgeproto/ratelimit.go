package edgeproto

import (
	fmt "fmt"
)

var GlobalApiName = "Global"

func (r *RateLimitSettings) Validate(fields map[string]struct{}) error {
	for _, fsettings := range r.FlowSettings {
		// Validate fields that must be set if FlowAlgorithm is set
		if fsettings.FlowAlgorithm == FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM || fsettings.FlowAlgorithm == FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM {
			if fsettings.ReqsPerSecond <= 0 {
				return fmt.Errorf("Invalid ReqsPerSecond %f, must be greater than 0", fsettings.ReqsPerSecond)
			}
			if fsettings.FlowAlgorithm == FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM {
				if fsettings.BurstSize <= 0 {
					return fmt.Errorf("Invalid BurstSize %d, must be greater than 0", fsettings.BurstSize)
				}
			}
		} else {
			return fmt.Errorf("Invalid FlowAlgorithm %v", fsettings.FlowAlgorithm)
		}
	}

	for _, msettings := range r.MaxReqsSettings {
		// Validate fields that must be set if MaxReqsAlgorithm is set
		if msettings.MaxReqsAlgorithm == MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM {
			if msettings.MaxRequests <= 0 {
				return fmt.Errorf("Invalid MaxRequests %d, must be greater than 0", msettings.MaxRequests)
			}
			if msettings.Interval <= 0 {
				return fmt.Errorf("Invalid Interval %d, must be greater than 0", msettings.Interval)
			}
		} else {
			return fmt.Errorf("Invalid MaxReqsAlgorithm %v", msettings.MaxReqsAlgorithm)
		}
	}
	return nil
}

// TODO: Consolidate with other validate function
func (f *FlowRateLimitSettings) Validate(fields map[string]struct{}) error {
	for field, _ := range fields {
		switch field {
		case FlowRateLimitSettingsFieldSettingsFlowAlgorithm:
			if f.Settings.FlowAlgorithm == FlowRateLimitAlgorithm_UNKNOWN_FLOW_ALGORITHM {
				return fmt.Errorf("Invalid FlowAlgorithm %v", f.Settings.FlowAlgorithm)
			}
		case FlowRateLimitSettingsFieldSettingsReqsPerSecond:
			if f.Settings.ReqsPerSecond <= 0 {
				return fmt.Errorf("Invalid ReqsPerSecond %f, must be greater than 0", f.Settings.ReqsPerSecond)
			}
		case FlowRateLimitSettingsFieldSettingsBurstSize:
			if f.Settings.FlowAlgorithm == FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM {
				if f.Settings.BurstSize <= 0 {
					return fmt.Errorf("Invalid BurstSize %d, must be greater than 0", f.Settings.BurstSize)
				}
			}
		}
	}
	return nil
}

func (m *MaxReqsRateLimitSettings) Validate(fields map[string]struct{}) error {
	for field, _ := range fields {
		switch field {
		case MaxReqsRateLimitSettingsFieldSettingsMaxReqsAlgorithm:
			if m.Settings.MaxReqsAlgorithm == MaxReqsRateLimitAlgorithm_UNKNOWN_MAX_REQS_ALGORITHM {
				return fmt.Errorf("Invalid MaxReqsAlgorithm %v", m.Settings.MaxReqsAlgorithm)
			}
		case MaxReqsRateLimitSettingsFieldSettingsMaxRequests:
			if m.Settings.MaxRequests <= 0 {
				return fmt.Errorf("Invalid MaxRequests %d, must be greater than 0", m.Settings.MaxRequests)
			}
		case MaxReqsRateLimitSettingsFieldSettingsInterval:
			if m.Settings.Interval <= 0 {
				return fmt.Errorf("Invalid Interval %d, must be greater than 0", m.Settings.Interval)
			}
		}
	}
	return nil
}

func (key *RateLimitSettingsKey) ValidateKey() error {
	if key == nil {
		return fmt.Errorf("Nil key")
	}
	if key.ApiName == "" {
		return fmt.Errorf("Invalid ApiName")
	}
	if key.RateLimitTarget == RateLimitTarget_UNKNOWN_TARGET {
		return fmt.Errorf("Invalid RateLimitTarget")
	}
	if key.ApiEndpointType == ApiEndpointType_UNKNOWN_API_ENDPOINT_TYPE {
		return fmt.Errorf("Invalid ApiEndpointType")
	}
	return nil
}

func (key *FlowRateLimitSettingsKey) ValidateKey() error {
	if key == nil {
		return fmt.Errorf("Nil key")
	}
	if key.FlowSettingsName == "" {
		return fmt.Errorf("Invalid FlowSettingsName")
	}
	return key.RateLimitKey.ValidateKey()
}

func (key *MaxReqsRateLimitSettingsKey) ValidateKey() error {
	if key == nil {
		return fmt.Errorf("Nil key")
	}
	if key.MaxReqsSettingsName == "" {
		return fmt.Errorf("Invalid MaxReqsSettingsName")
	}
	return key.RateLimitKey.ValidateKey()
}

func GetRateLimitSettingsKey(apiName string, apiEndpointType ApiEndpointType, rateLimitTarget RateLimitTarget) RateLimitSettingsKey {
	return RateLimitSettingsKey{
		ApiName:         apiName,
		ApiEndpointType: apiEndpointType,
		RateLimitTarget: rateLimitTarget,
	}
}

// Returns map of Default RateLimitSettings
func GetDefaultRateLimitSettings() map[RateLimitSettingsKey]*RateLimitSettings {
	// Init all AllRequests RateLimitSettings
	dmeGlobalAllReqs := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_DME,
			RateLimitTarget: RateLimitTarget_ALL_REQUESTS,
			ApiName:         GlobalApiName,
		},
		FlowSettings: map[string]*FlowSettings{
			"dmeglobalallreqs1": &FlowSettings{
				FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
				ReqsPerSecond: 25000,
				BurstSize:     250,
			},
		},
	}
	verifyLocAllReqs := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_DME,
			RateLimitTarget: RateLimitTarget_ALL_REQUESTS,
			ApiName:         "VerifyLocation",
		},
		FlowSettings: map[string]*FlowSettings{
			"verifylocallreqs1": &FlowSettings{
				FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
				ReqsPerSecond: 5000,
				BurstSize:     50,
			},
		},
	}
	// Init all PerIp RateLimitSettings
	dmeGlobalPerIp := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_DME,
			RateLimitTarget: RateLimitTarget_PER_IP,
			ApiName:         GlobalApiName,
		},
		FlowSettings: map[string]*FlowSettings{
			"dmeglobalperip1": &FlowSettings{
				FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
				ReqsPerSecond: 10000,
				BurstSize:     100,
			},
		},
	}
	verifyLocPerIp := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_DME,
			RateLimitTarget: RateLimitTarget_PER_IP,
			ApiName:         "VerifyLocation",
		},
		FlowSettings: map[string]*FlowSettings{
			"verifylocperip1": &FlowSettings{
				FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
				ReqsPerSecond: 1000,
				BurstSize:     25,
			},
		},
	}
	// Assign RateLimitSettings to RateLimitSettingsKey
	rlMap := make(map[RateLimitSettingsKey]*RateLimitSettings)
	rlMap[dmeGlobalAllReqs.Key] = dmeGlobalAllReqs
	rlMap[verifyLocAllReqs.Key] = verifyLocAllReqs
	rlMap[dmeGlobalPerIp.Key] = dmeGlobalPerIp
	rlMap[verifyLocPerIp.Key] = verifyLocPerIp

	return rlMap
}
