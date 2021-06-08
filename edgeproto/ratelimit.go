package edgeproto

import (
	fmt "fmt"
)

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

func (key *RateLimitSettingsKey) ValidateKey() error {
	if key == nil {
		return fmt.Errorf("Nil key")
	}
	if key.ApiEndpointType == ApiEndpointType_UNKNOWN_API_ENDPOINT_TYPE {
		return fmt.Errorf("Invalid ApiEndpointType")
	}
	if key.RateLimitTarget == RateLimitTarget_UNKNOWN_TARGET {
		return fmt.Errorf("Invalid RateLimitTarget")
	}
	return nil
}

func GetRateLimitSettingsKey(apiEndpointType ApiEndpointType, rateLimitTarget RateLimitTarget, apiName string) RateLimitSettingsKey {
	return RateLimitSettingsKey{
		ApiEndpointType: apiEndpointType,
		RateLimitTarget: rateLimitTarget,
		ApiName:         apiName,
	}
}

// Returns map of Default DME RateLimitSettings
func GetDefaultDmeRateLimitSettings() map[RateLimitSettingsKey]*RateLimitSettings {
	// Init all AllRequests RateLimitSettings
	dmeDefaultAllReqs := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_DME,
			ApiActionType:   ApiActionType_DEFAULT_ACTION,
			RateLimitTarget: RateLimitTarget_ALL_REQUESTS,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 50,
		BurstSize:     5,
	}
	// Init all PerIp RateLimitSettings
	dmeDefaultPerIp := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_DME,
			ApiActionType:   ApiActionType_DEFAULT_ACTION,
			RateLimitTarget: RateLimitTarget_PER_IP,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 10,
		BurstSize:     3,
	}
	// Assign RateLimitSettings to RateLimitSettingsKey
	rlMap := make(map[RateLimitSettingsKey]*RateLimitSettings)
	rlMap[dmeDefaultAllReqs.Key] = dmeDefaultAllReqs
	rlMap[dmeDefaultPerIp.Key] = dmeDefaultPerIp

	return rlMap
}

// Returns map of Default Controller RateLimitSettings
func GetDefaultControllerRateLimitSettings() map[RateLimitSettingsKey]*RateLimitSettings {
	// Init all AllRequests RateLimitSettings
	ctrlCreateAllReqs := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_CONTROLLER,
			ApiActionType:   ApiActionType_CREATE_ACTION,
			RateLimitTarget: RateLimitTarget_ALL_REQUESTS,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 50,
		BurstSize:     5,
	}
	ctrlDeleteAllReqs := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_CONTROLLER,
			ApiActionType:   ApiActionType_DELETE_ACTION,
			RateLimitTarget: RateLimitTarget_ALL_REQUESTS,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 50,
		BurstSize:     5,
	}
	ctrlUpdateAllReqs := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_CONTROLLER,
			ApiActionType:   ApiActionType_UPDATE_ACTION,
			RateLimitTarget: RateLimitTarget_ALL_REQUESTS,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 50,
		BurstSize:     5,
	}
	ctrlShowAllReqs := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_CONTROLLER,
			ApiActionType:   ApiActionType_SHOW_ACTION,
			RateLimitTarget: RateLimitTarget_ALL_REQUESTS,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 100,
		BurstSize:     10,
	}
	ctrlDefaultAllReqs := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_CONTROLLER,
			ApiActionType:   ApiActionType_DEFAULT_ACTION,
			RateLimitTarget: RateLimitTarget_ALL_REQUESTS,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 50,
		BurstSize:     5,
	}
	// Init all PerIp RateLimitSettings
	ctrlCreatePerIp := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_CONTROLLER,
			ApiActionType:   ApiActionType_CREATE_ACTION,
			RateLimitTarget: RateLimitTarget_PER_IP,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 10,
		BurstSize:     3,
	}
	ctrlDeletePerIp := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_CONTROLLER,
			ApiActionType:   ApiActionType_DELETE_ACTION,
			RateLimitTarget: RateLimitTarget_PER_IP,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 10,
		BurstSize:     3,
	}
	ctrlUpdatePerIp := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_CONTROLLER,
			ApiActionType:   ApiActionType_UPDATE_ACTION,
			RateLimitTarget: RateLimitTarget_PER_IP,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 10,
		BurstSize:     3,
	}
	ctrlShowPerIp := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_CONTROLLER,
			ApiActionType:   ApiActionType_SHOW_ACTION,
			RateLimitTarget: RateLimitTarget_PER_IP,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 15,
		BurstSize:     3,
	}
	ctrlDefaultPerIp := &RateLimitSettings{
		Key: RateLimitSettingsKey{
			ApiEndpointType: ApiEndpointType_CONTROLLER,
			ApiActionType:   ApiActionType_DEFAULT_ACTION,
			RateLimitTarget: RateLimitTarget_PER_IP,
		},
		FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
		ReqsPerSecond: 10,
		BurstSize:     3,
	}
	// Assign RateLimitSettings to RateLimitSettingsKey
	rlMap := make(map[RateLimitSettingsKey]*RateLimitSettings)
	rlMap[ctrlCreateAllReqs.Key] = ctrlCreateAllReqs
	rlMap[ctrlDeleteAllReqs.Key] = ctrlDeleteAllReqs
	rlMap[ctrlUpdateAllReqs.Key] = ctrlUpdateAllReqs
	rlMap[ctrlShowAllReqs.Key] = ctrlShowAllReqs
	rlMap[ctrlDefaultAllReqs.Key] = ctrlDefaultAllReqs
	rlMap[ctrlCreatePerIp.Key] = ctrlCreatePerIp
	rlMap[ctrlDeletePerIp.Key] = ctrlDeletePerIp
	rlMap[ctrlUpdatePerIp.Key] = ctrlUpdatePerIp
	rlMap[ctrlShowPerIp.Key] = ctrlShowPerIp
	rlMap[ctrlDefaultPerIp.Key] = ctrlDefaultPerIp

	return rlMap
}
