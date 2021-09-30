package ratelimit

import (
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/require"
)

var flowSettings0 = edgeproto.FlowSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 25000,
	BurstSize:     250,
}

var flowSettings1 = edgeproto.FlowSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 5000,
	BurstSize:     50,
}

var flowSettings2 = edgeproto.FlowSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 10000,
	BurstSize:     1000,
}

var flowSettings3 = edgeproto.FlowSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM,
	ReqsPerSecond: 20000,
	BurstSize:     2000,
}

var flowSettings4 = edgeproto.FlowSettings{
	FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
	ReqsPerSecond: 1000,
	BurstSize:     25,
}

var maxSettings0 = edgeproto.MaxReqsSettings{
	MaxReqsAlgorithm: edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM,
	MaxRequests:      10000,
	Interval:         edgeproto.Duration(time.Second),
}

var maxSettings1 = edgeproto.MaxReqsSettings{
	MaxReqsAlgorithm: edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM,
	MaxRequests:      20000,
	Interval:         edgeproto.Duration(time.Second),
}

// Create copies so global values will not be modified by tests.
func newFlowSettingsX(ii int) *edgeproto.FlowSettings {
	f := edgeproto.FlowSettings{}
	switch ii {
	case 0:
		f = flowSettings0
	case 1:
		f = flowSettings1
	case 2:
		f = flowSettings2
	case 3:
		f = flowSettings3
	case 4:
		f = flowSettings4
	}
	return &f
}

// Create copies so global values will not be modified by tests.
func newMaxSettingsX(ii int) *edgeproto.MaxReqsSettings {
	m := edgeproto.MaxReqsSettings{}
	switch ii {
	case 0:
		m = maxSettings0
	case 1:
		m = maxSettings1
	}
	return &m
}

var userSettings = []*edgeproto.RateLimitSettings{
	&edgeproto.RateLimitSettings{
		Key: edgeproto.RateLimitSettingsKey{
			ApiEndpointType: edgeproto.ApiEndpointType_DME,
			RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
			ApiName:         edgeproto.GlobalApiName,
		},
		FlowSettings: map[string]*edgeproto.FlowSettings{
			"dmeglobalallreqs0": newFlowSettingsX(0),
			"leakyallreqs3":     newFlowSettingsX(3),
		},
		MaxReqsSettings: map[string]*edgeproto.MaxReqsSettings{
			"dmeglobalmax0": newMaxSettingsX(0),
		},
	},
	&edgeproto.RateLimitSettings{
		Key: edgeproto.RateLimitSettingsKey{
			ApiEndpointType: edgeproto.ApiEndpointType_DME,
			RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
			ApiName:         "VerifyLocation",
		},
		FlowSettings: map[string]*edgeproto.FlowSettings{
			"verifylocalreqs1": newFlowSettingsX(1),
		},
	},
	&edgeproto.RateLimitSettings{
		Key: edgeproto.RateLimitSettingsKey{
			ApiEndpointType: edgeproto.ApiEndpointType_DME,
			RateLimitTarget: edgeproto.RateLimitTarget_PER_IP,
			ApiName:         edgeproto.GlobalApiName,
		},
		FlowSettings: map[string]*edgeproto.FlowSettings{
			"dmeglobalperip0": newFlowSettingsX(0),
			"dmeglobalperip2": newFlowSettingsX(2),
		},
		MaxReqsSettings: map[string]*edgeproto.MaxReqsSettings{
			"dmeglobalmax1": newMaxSettingsX(1),
		},
	},
	&edgeproto.RateLimitSettings{
		Key: edgeproto.RateLimitSettingsKey{
			ApiEndpointType: edgeproto.ApiEndpointType_DME,
			RateLimitTarget: edgeproto.RateLimitTarget_PER_IP,
			ApiName:         "VerifyLocation",
		},
		FlowSettings: map[string]*edgeproto.FlowSettings{
			"verifylocperip4": newFlowSettingsX(4),
		},
	},
}

var dbFlowSettings = []*edgeproto.FlowRateLimitSettings{
	&edgeproto.FlowRateLimitSettings{
		Key: edgeproto.FlowRateLimitSettingsKey{
			FlowSettingsName: "dmeglobalallreqs0",
			RateLimitKey:     userSettings[0].Key,
		},
		Settings: flowSettings0,
	},
	&edgeproto.FlowRateLimitSettings{
		Key: edgeproto.FlowRateLimitSettingsKey{
			FlowSettingsName: "leakyallreqs3",
			RateLimitKey:     userSettings[0].Key,
		},
		Settings: flowSettings3,
	},
	&edgeproto.FlowRateLimitSettings{
		Key: edgeproto.FlowRateLimitSettingsKey{
			FlowSettingsName: "verifylocalreqs1",
			RateLimitKey:     userSettings[1].Key,
		},
		Settings: flowSettings1,
	},
	&edgeproto.FlowRateLimitSettings{
		Key: edgeproto.FlowRateLimitSettingsKey{
			FlowSettingsName: "dmeglobalperip0",
			RateLimitKey:     userSettings[2].Key,
		},
		Settings: flowSettings0,
	},
	&edgeproto.FlowRateLimitSettings{
		Key: edgeproto.FlowRateLimitSettingsKey{
			FlowSettingsName: "dmeglobalperip2",
			RateLimitKey:     userSettings[2].Key,
		},
		Settings: flowSettings2,
	},
	&edgeproto.FlowRateLimitSettings{
		Key: edgeproto.FlowRateLimitSettingsKey{
			FlowSettingsName: "verifylocperip4",
			RateLimitKey:     userSettings[3].Key,
		},
		Settings: flowSettings4,
	},
}

var dbMaxSettings = []*edgeproto.MaxReqsRateLimitSettings{
	&edgeproto.MaxReqsRateLimitSettings{
		Key: edgeproto.MaxReqsRateLimitSettingsKey{
			MaxReqsSettingsName: "dmeglobalmax0",
			RateLimitKey:        userSettings[0].Key,
		},
		Settings: maxSettings0,
	},
	&edgeproto.MaxReqsRateLimitSettings{
		Key: edgeproto.MaxReqsRateLimitSettingsKey{
			MaxReqsSettingsName: "dmeglobalmax1",
			RateLimitKey:        userSettings[2].Key,
		},
		Settings: maxSettings1,
	},
}

func TestConversion(t *testing.T) {
	dbf, dbm := UserToDbSettings(userSettings)
	require.Equal(t, dbFlowSettings, dbf)
	require.Equal(t, dbMaxSettings, dbm)
	sets := DbToUserSettings(dbFlowSettings, dbMaxSettings)
	require.Equal(t, userSettings, sets)
}
