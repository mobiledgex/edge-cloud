// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ratelimit.proto

package gencmd

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var ApiEndpointRateLimitSettingsRequiredArgs = []string{}
var ApiEndpointRateLimitSettingsOptionalArgs = []string{
	"removeratelimit",
	"endpointratelimitsettings.flowratelimitsettings.flowalgorithm",
	"endpointratelimitsettings.flowratelimitsettings.reqspersecond",
	"endpointratelimitsettings.flowratelimitsettings.burstsize",
	"endpointratelimitsettings.maxreqsratelimitsettings.maxreqsalgorithm",
	"endpointratelimitsettings.maxreqsratelimitsettings.maxrequestspersecond",
	"endpointratelimitsettings.maxreqsratelimitsettings.maxrequestsperminute",
	"endpointratelimitsettings.maxreqsratelimitsettings.maxrequestsperhour",
	"endpointperipratelimitsettings.flowratelimitsettings.flowalgorithm",
	"endpointperipratelimitsettings.flowratelimitsettings.reqspersecond",
	"endpointperipratelimitsettings.flowratelimitsettings.burstsize",
	"endpointperipratelimitsettings.maxreqsratelimitsettings.maxreqsalgorithm",
	"endpointperipratelimitsettings.maxreqsratelimitsettings.maxrequestspersecond",
	"endpointperipratelimitsettings.maxreqsratelimitsettings.maxrequestsperminute",
	"endpointperipratelimitsettings.maxreqsratelimitsettings.maxrequestsperhour",
	"endpointperuserratelimitsettings.flowratelimitsettings.flowalgorithm",
	"endpointperuserratelimitsettings.flowratelimitsettings.reqspersecond",
	"endpointperuserratelimitsettings.flowratelimitsettings.burstsize",
	"endpointperuserratelimitsettings.maxreqsratelimitsettings.maxreqsalgorithm",
	"endpointperuserratelimitsettings.maxreqsratelimitsettings.maxrequestspersecond",
	"endpointperuserratelimitsettings.maxreqsratelimitsettings.maxrequestsperminute",
	"endpointperuserratelimitsettings.maxreqsratelimitsettings.maxrequestsperhour",
	"endpointperorgratelimitsettings.flowratelimitsettings.flowalgorithm",
	"endpointperorgratelimitsettings.flowratelimitsettings.reqspersecond",
	"endpointperorgratelimitsettings.flowratelimitsettings.burstsize",
	"endpointperorgratelimitsettings.maxreqsratelimitsettings.maxreqsalgorithm",
	"endpointperorgratelimitsettings.maxreqsratelimitsettings.maxrequestspersecond",
	"endpointperorgratelimitsettings.maxreqsratelimitsettings.maxrequestsperminute",
	"endpointperorgratelimitsettings.maxreqsratelimitsettings.maxrequestsperhour",
}
var ApiEndpointRateLimitSettingsAliasArgs = []string{}
var ApiEndpointRateLimitSettingsComments = map[string]string{
	"removeratelimit": "If set to true, no rate limiting will occur",
	"endpointratelimitsettings.flowratelimitsettings.flowalgorithm":                  "Flow Rate Limit algorithm - includes NoFlowAlgorithm, TokenBucketAlgorithm, or LeakyBucketAlgorithm, one of NoFlowAlgorithm, TokenBucketAlgorithm, LeakyBucketAlgorithm",
	"endpointratelimitsettings.flowratelimitsettings.reqspersecond":                  "requests per second",
	"endpointratelimitsettings.flowratelimitsettings.burstsize":                      "burst size",
	"endpointratelimitsettings.maxreqsratelimitsettings.maxreqsalgorithm":            "MaxReqs Rate Limit Algorithm - includes NoMaxReqsAlgorithm or FixedWindowAlgorithm, one of NoMaxReqsAlgorithm, FixedWindowAlgorithm, RollingWindowAlgorithm",
	"endpointratelimitsettings.maxreqsratelimitsettings.maxrequestspersecond":        "maximum number of requests per second",
	"endpointratelimitsettings.maxreqsratelimitsettings.maxrequestsperminute":        "maximum number of requests per minute",
	"endpointratelimitsettings.maxreqsratelimitsettings.maxrequestsperhour":          "maximum number of requests per hour",
	"endpointperipratelimitsettings.flowratelimitsettings.flowalgorithm":             "Flow Rate Limit algorithm - includes NoFlowAlgorithm, TokenBucketAlgorithm, or LeakyBucketAlgorithm, one of NoFlowAlgorithm, TokenBucketAlgorithm, LeakyBucketAlgorithm",
	"endpointperipratelimitsettings.flowratelimitsettings.reqspersecond":             "requests per second",
	"endpointperipratelimitsettings.flowratelimitsettings.burstsize":                 "burst size",
	"endpointperipratelimitsettings.maxreqsratelimitsettings.maxreqsalgorithm":       "MaxReqs Rate Limit Algorithm - includes NoMaxReqsAlgorithm or FixedWindowAlgorithm, one of NoMaxReqsAlgorithm, FixedWindowAlgorithm, RollingWindowAlgorithm",
	"endpointperipratelimitsettings.maxreqsratelimitsettings.maxrequestspersecond":   "maximum number of requests per second",
	"endpointperipratelimitsettings.maxreqsratelimitsettings.maxrequestsperminute":   "maximum number of requests per minute",
	"endpointperipratelimitsettings.maxreqsratelimitsettings.maxrequestsperhour":     "maximum number of requests per hour",
	"endpointperuserratelimitsettings.flowratelimitsettings.flowalgorithm":           "Flow Rate Limit algorithm - includes NoFlowAlgorithm, TokenBucketAlgorithm, or LeakyBucketAlgorithm, one of NoFlowAlgorithm, TokenBucketAlgorithm, LeakyBucketAlgorithm",
	"endpointperuserratelimitsettings.flowratelimitsettings.reqspersecond":           "requests per second",
	"endpointperuserratelimitsettings.flowratelimitsettings.burstsize":               "burst size",
	"endpointperuserratelimitsettings.maxreqsratelimitsettings.maxreqsalgorithm":     "MaxReqs Rate Limit Algorithm - includes NoMaxReqsAlgorithm or FixedWindowAlgorithm, one of NoMaxReqsAlgorithm, FixedWindowAlgorithm, RollingWindowAlgorithm",
	"endpointperuserratelimitsettings.maxreqsratelimitsettings.maxrequestspersecond": "maximum number of requests per second",
	"endpointperuserratelimitsettings.maxreqsratelimitsettings.maxrequestsperminute": "maximum number of requests per minute",
	"endpointperuserratelimitsettings.maxreqsratelimitsettings.maxrequestsperhour":   "maximum number of requests per hour",
	"endpointperorgratelimitsettings.flowratelimitsettings.flowalgorithm":            "Flow Rate Limit algorithm - includes NoFlowAlgorithm, TokenBucketAlgorithm, or LeakyBucketAlgorithm, one of NoFlowAlgorithm, TokenBucketAlgorithm, LeakyBucketAlgorithm",
	"endpointperorgratelimitsettings.flowratelimitsettings.reqspersecond":            "requests per second",
	"endpointperorgratelimitsettings.flowratelimitsettings.burstsize":                "burst size",
	"endpointperorgratelimitsettings.maxreqsratelimitsettings.maxreqsalgorithm":      "MaxReqs Rate Limit Algorithm - includes NoMaxReqsAlgorithm or FixedWindowAlgorithm, one of NoMaxReqsAlgorithm, FixedWindowAlgorithm, RollingWindowAlgorithm",
	"endpointperorgratelimitsettings.maxreqsratelimitsettings.maxrequestspersecond":  "maximum number of requests per second",
	"endpointperorgratelimitsettings.maxreqsratelimitsettings.maxrequestsperminute":  "maximum number of requests per minute",
	"endpointperorgratelimitsettings.maxreqsratelimitsettings.maxrequestsperhour":    "maximum number of requests per hour",
}
var ApiEndpointRateLimitSettingsSpecialArgs = map[string]string{}
var RateLimitSettingsRequiredArgs = []string{}
var RateLimitSettingsOptionalArgs = []string{
	"flowratelimitsettings.flowalgorithm",
	"flowratelimitsettings.reqspersecond",
	"flowratelimitsettings.burstsize",
	"maxreqsratelimitsettings.maxreqsalgorithm",
	"maxreqsratelimitsettings.maxrequestspersecond",
	"maxreqsratelimitsettings.maxrequestsperminute",
	"maxreqsratelimitsettings.maxrequestsperhour",
}
var RateLimitSettingsAliasArgs = []string{}
var RateLimitSettingsComments = map[string]string{
	"flowratelimitsettings.flowalgorithm":           "Flow Rate Limit algorithm - includes NoFlowAlgorithm, TokenBucketAlgorithm, or LeakyBucketAlgorithm, one of NoFlowAlgorithm, TokenBucketAlgorithm, LeakyBucketAlgorithm",
	"flowratelimitsettings.reqspersecond":           "requests per second",
	"flowratelimitsettings.burstsize":               "burst size",
	"maxreqsratelimitsettings.maxreqsalgorithm":     "MaxReqs Rate Limit Algorithm - includes NoMaxReqsAlgorithm or FixedWindowAlgorithm, one of NoMaxReqsAlgorithm, FixedWindowAlgorithm, RollingWindowAlgorithm",
	"maxreqsratelimitsettings.maxrequestspersecond": "maximum number of requests per second",
	"maxreqsratelimitsettings.maxrequestsperminute": "maximum number of requests per minute",
	"maxreqsratelimitsettings.maxrequestsperhour":   "maximum number of requests per hour",
}
var RateLimitSettingsSpecialArgs = map[string]string{}
var FlowRateLimitSettingsRequiredArgs = []string{}
var FlowRateLimitSettingsOptionalArgs = []string{
	"flowalgorithm",
	"reqspersecond",
	"burstsize",
}
var FlowRateLimitSettingsAliasArgs = []string{}
var FlowRateLimitSettingsComments = map[string]string{
	"flowalgorithm": "Flow Rate Limit algorithm - includes NoFlowAlgorithm, TokenBucketAlgorithm, or LeakyBucketAlgorithm, one of NoFlowAlgorithm, TokenBucketAlgorithm, LeakyBucketAlgorithm",
	"reqspersecond": "requests per second",
	"burstsize":     "burst size",
}
var FlowRateLimitSettingsSpecialArgs = map[string]string{}
var MaxReqsRateLimitSettingsRequiredArgs = []string{}
var MaxReqsRateLimitSettingsOptionalArgs = []string{
	"maxreqsalgorithm",
	"maxrequestspersecond",
	"maxrequestsperminute",
	"maxrequestsperhour",
}
var MaxReqsRateLimitSettingsAliasArgs = []string{}
var MaxReqsRateLimitSettingsComments = map[string]string{
	"maxreqsalgorithm":     "MaxReqs Rate Limit Algorithm - includes NoMaxReqsAlgorithm or FixedWindowAlgorithm, one of NoMaxReqsAlgorithm, FixedWindowAlgorithm, RollingWindowAlgorithm",
	"maxrequestspersecond": "maximum number of requests per second",
	"maxrequestsperminute": "maximum number of requests per minute",
	"maxrequestsperhour":   "maximum number of requests per hour",
}
var MaxReqsRateLimitSettingsSpecialArgs = map[string]string{}
