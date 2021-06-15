package ratelimit

import (
	"context"
)

/*
 * Limiter Interface
 * Structs that implement this inferface must provide a limit function that returns whether or not to allow a request to go through
 * Return value is an error
 * If the return value is non-nil, we will reject the request (ie. limit), a return value of nil will pass the request.
 * Current implementations in: api_ratelimitmgr.go, apiendpoint-limiter.go, leakybucket.go, tokenbucket.go
 */
type Limiter interface {
	Limit(ctx context.Context, info *CallerInfo) error
	Type() string
}

// Struct used to supply client/caller information to Limiters
type CallerInfo struct {
	Api  string
	User string
	Ip   string
}

var DefaultReqsPerSecondPerApi = 100.0
var DefaultTokenBucketSize int64 = 10 // equivalent to burst size
