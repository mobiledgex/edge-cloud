package edgeproto

import fmt "fmt"

func (r *RateLimitSettings) Validate(fields map[string]struct{}) error {
	var err error
	if err = r.GetKey().ValidateKey(); err != nil {
		return err
	}
	return err
}

func (key *RateLimitSettingsKey) ValidateKey() error {
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

func GetDefaultRateLimitSettings() map[RateLimitSettingsKey]*RateLimitSettings {
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

	rlMap := make(map[RateLimitSettingsKey]*RateLimitSettings)
	rlMap[ctrlCreateAllReqs.Key] = ctrlCreateAllReqs
	rlMap[ctrlDeleteAllReqs.Key] = ctrlDeleteAllReqs
	rlMap[ctrlUpdateAllReqs.Key] = ctrlUpdateAllReqs
	rlMap[ctrlShowAllReqs.Key] = ctrlShowAllReqs
	rlMap[ctrlDefaultAllReqs.Key] = ctrlDefaultAllReqs
	rlMap[dmeDefaultAllReqs.Key] = dmeDefaultAllReqs
	rlMap[ctrlCreatePerIp.Key] = ctrlCreatePerIp
	rlMap[ctrlDeletePerIp.Key] = ctrlDeletePerIp
	rlMap[ctrlUpdatePerIp.Key] = ctrlUpdatePerIp
	rlMap[ctrlShowPerIp.Key] = ctrlShowPerIp
	rlMap[ctrlDefaultPerIp.Key] = ctrlDefaultPerIp
	rlMap[dmeDefaultPerIp.Key] = dmeDefaultPerIp

	return rlMap
}
