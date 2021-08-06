package edgeproto

import (
	fmt "fmt"
)

var GlobalApiName = "Global"

// TODO: VALIDATE ALL BASED ON FIELDS
func (f *FlowSettings) Validate() error {
	// Validate fields that must be set if FlowAlgorithm is set
	if f.FlowAlgorithm == FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM || f.FlowAlgorithm == FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM {
		if f.ReqsPerSecond <= 0 {
			return fmt.Errorf("Invalid ReqsPerSecond %f, must be greater than 0", f.ReqsPerSecond)
		}
		if f.FlowAlgorithm == FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM {
			if f.BurstSize <= 0 {
				return fmt.Errorf("Invalid BurstSize %d, must be greater than 0", f.BurstSize)
			}
		}
	} else {
		return fmt.Errorf("Invalid FlowAlgorithm %v", f.FlowAlgorithm)
	}
	return nil
}

func (m *MaxReqsSettings) Validate() error {
	// Validate fields that must be set if MaxReqsAlgorithm is set
	if m.MaxReqsAlgorithm == MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM {
		if m.MaxRequests <= 0 {
			return fmt.Errorf("Invalid MaxRequests %d, must be greater than 0", m.MaxRequests)
		}
		if m.Interval <= 0 {
			return fmt.Errorf("Invalid Interval %d, must be greater than 0", m.Interval)
		}
	} else {
		return fmt.Errorf("Invalid MaxReqsAlgorithm %v", m.MaxReqsAlgorithm)
	}
	return nil
}

func (r *RateLimitSettings) Validate(fields map[string]struct{}) error {
	for _, fsettings := range r.FlowSettings {
		if err := fsettings.Validate(); err != nil {
			return err
		}
	}

	for _, msettings := range r.MaxReqsSettings {
		if err := msettings.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (f *FlowRateLimitSettings) Validate(fields map[string]struct{}) error {
	return f.Settings.Validate()
}

func (m *MaxReqsRateLimitSettings) Validate(fields map[string]struct{}) error {
	return m.Settings.Validate()
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

func (r *RateLimitSettings) UpdateFlowSettings(f *FlowRateLimitSettings) {
	if r.FlowSettings == nil {
		r.FlowSettings = make(map[string]*FlowSettings)
	}
	r.FlowSettings[f.Key.FlowSettingsName] = &f.Settings
}

func (r *RateLimitSettings) RemoveFlowSettings(name string) {
	delete(r.FlowSettings, name)
}

func (r *RateLimitSettings) UpdateMaxReqsSettings(m *MaxReqsRateLimitSettings) {
	if r.MaxReqsSettings == nil {
		r.MaxReqsSettings = make(map[string]*MaxReqsSettings)
	}
	r.MaxReqsSettings[m.Key.MaxReqsSettingsName] = &m.Settings
}

func (r *RateLimitSettings) RemoveMaxReqsSettings(name string) {
	delete(r.MaxReqsSettings, name)
}
