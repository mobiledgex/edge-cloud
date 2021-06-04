package edgeproto

import (
	fmt "fmt"
)

func (r *RateLimitSettings) Validate(fields map[string]struct{}) error {
	// Validate fields that must be set if FlowAlgorithm is set
	if r.FlowAlgorithm == FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM || r.FlowAlgorithm == FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM {
		if r.ReqsPerSecond <= 0 {
			return fmt.Errorf("Invalid ReqsPerSecond %f, must be greater than 0", r.ReqsPerSecond)
		}
		if r.FlowAlgorithm == FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM {
			if r.BurstSize <= 0 {
				return fmt.Errorf("Invalid BurstSize %d, must be greater than 0", r.BurstSize)
			}
		}
	}

	// RollingWindowAlgorithm is not implemented yet
	if r.MaxReqsAlgorithm == MaxReqsRateLimitAlgorithm_ROLLING_WINDOW_ALGORITHM {
		return fmt.Errorf("Invalid MaxReqsRateLimitAlgorithm %v, only FixedWindowAlgorithm is implemented", r.MaxReqsAlgorithm)
	}

	// Validate fields that must be set if MaxReqsAlgorithm is set
	if r.MaxReqsAlgorithm == MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM {
		if r.MaxRequestsPerSecond <= 0 && r.MaxRequestsPerMinute <= 0 && r.MaxRequestsPerHour <= 0 {
			return fmt.Errorf("One of MaxRequestsPerSecond, MaxRequestsPerMinute, or MaxRequestsPerHour must be greater than 0 to use MaxReqs limiting")
		}
	}

	// Validate fields that must be set if ReqsPerSecond is set
	if r.ReqsPerSecond != 0 {
		if r.ReqsPerSecond < 0 {
			return fmt.Errorf("Invalid ReqsPerSecond %f, must be greater than 0", r.ReqsPerSecond)
		}
		if r.FlowAlgorithm != FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM && r.FlowAlgorithm != FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM {
			return fmt.Errorf("Must have valid FlowRateLimitAlgorithm if ReqsPerSecond is set")
		}
	}

	// Validate fields that must be set if BurstSize is set
	if r.BurstSize != 0 {
		if r.BurstSize < 0 {
			return fmt.Errorf("Invalid BurstSize %d, must be greater than 0", r.BurstSize)
		}
		if r.FlowAlgorithm != FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM && r.FlowAlgorithm != FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM {
			return fmt.Errorf("Must have valid FlowRateLimitAlgorithm if BurstSize is set")
		}
	}

	// Validate fields that must be set if MaxRequestsPerSecond is set
	if r.MaxRequestsPerSecond != 0 {
		if r.MaxRequestsPerSecond < 0 {
			return fmt.Errorf("Invalid MaxRequestsPerSecond %d, must be greater than 0", r.MaxRequestsPerSecond)
		}
		if r.MaxReqsAlgorithm != MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM {
			return fmt.Errorf("Must have valid MaxReqsRateLimitAlgorithm if MaxRequestsPerSecond is set")
		}
	}

	// Validate fields that must be set if MaxRequestsPerMinute is set
	if r.MaxRequestsPerMinute != 0 {
		if r.MaxRequestsPerMinute < 0 {
			return fmt.Errorf("Invalid MaxRequestsPerMinute %d, must be greater than 0", r.MaxRequestsPerMinute)
		}
		if r.MaxReqsAlgorithm != MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM {
			return fmt.Errorf("Must have valid MaxReqsRateLimitAlgorithm if MaxRequestsPerMinute is set")
		}
	}

	// Validate fields that must be set if MaxRequestsPerHour is set
	if r.MaxRequestsPerHour != 0 {
		if r.MaxRequestsPerHour < 0 {
			return fmt.Errorf("Invalid MaxRequestsPerHour %d, must be greater than 0", r.MaxRequestsPerHour)
		}
		if r.MaxReqsAlgorithm != MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM {
			return fmt.Errorf("Must have valid MaxReqsRateLimitAlgorithm if MaxRequestsPerHour is set")
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
	if key.ApiActionType == ApiActionType_UNKNOWN_ACTION {
		return fmt.Errorf("Invalid ApiActionType")
	}
	if key.RateLimitTarget == RateLimitTarget_UNKNOWN_TARGET {
		return fmt.Errorf("Invalid RateLimitTarget")
	}
	return nil
}

func GetRateLimitSettingsKey(apiEndpointType ApiEndpointType, apiActionType ApiActionType, rateLimitTarget RateLimitTarget) RateLimitSettingsKey {
	return RateLimitSettingsKey{
		ApiEndpointType: apiEndpointType,
		ApiActionType:   apiActionType,
		RateLimitTarget: rateLimitTarget,
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
