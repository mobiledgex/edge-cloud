package edgeproto

func validateApiEndpointRateLimitSettings(v *FieldValidator, field string, s *ApiEndpointRateLimitSettings) {
	validateRateLimitSettings(v, field, s.EndpointRateLimitSettings)
	validateRateLimitSettings(v, field, s.EndpointPerIpRateLimitSettings)
	validateRateLimitSettings(v, field, s.EndpointPerUserRateLimitSettings)
	validateRateLimitSettings(v, field, s.EndpointPerOrgRateLimitSettings)
}

func validateRateLimitSettings(v *FieldValidator, field string, r *RateLimitSettings) {
	if r != nil {
		validateFlowRateLimitSettings(v, field, r.FlowRateLimitSettings)
		validateMaxReqsRateLimitSettings(v, field, r.MaxReqsRateLimitSettings)
	}
}

func validateFlowRateLimitSettings(v *FieldValidator, field string, f *FlowRateLimitSettings) {
	if f != nil {
		v.CheckGT(field, f.ReqsPerSecond, 0)
		v.CheckGT(field, f.BurstSize, 0)
	}
}

func validateMaxReqsRateLimitSettings(v *FieldValidator, field string, m *MaxReqsRateLimitSettings) {
	if m != nil {
		v.CheckGT(field, m.MaxRequestsPerSecond, 0)
		v.CheckGT(field, m.MaxRequestsPerMinute, 0)
		v.CheckGT(field, m.MaxRequestsPerHour, 0)
	}
}

var DefaultControllerCreateApiEndpointRateLimitSettings = &ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 5,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultControllerDeleteApiEndpointRateLimitSettings = &ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 5,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultControllerUpdateApiEndpointRateLimitSettings = &ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 5,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultControllerDefaultApiEndpointRateLimitSettings = &ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 5,
			BurstSize:     1,
		},
	},
	EndpointPerIpRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultControllerShowApiEndpointRateLimitSettings = &ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 10,
			BurstSize:     3,
		},
	},
	EndpointPerIpRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 1,
			BurstSize:     1,
		},
	},
}

var DefaultDmeDefaultApiEndpointRateLimitSettings = &ApiEndpointRateLimitSettings{
	EndpointRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 20,
			BurstSize:     5,
		},
	},
	EndpointPerIpRateLimitSettings: &RateLimitSettings{
		FlowRateLimitSettings: &FlowRateLimitSettings{
			FlowAlgorithm: FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
			ReqsPerSecond: 3,
			BurstSize:     1,
		},
	},
}

var DefaultReqsPerSecondPerApi = 100.0
var DefaultTokenBucketSize int64 = 10 // equivalent to burst size
