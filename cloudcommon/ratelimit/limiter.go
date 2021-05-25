package ratelimit

import (
	"context"
)

// Limiter Interface
// Structs that implement this inferface must provide a limit function that returns whether or not to allow a request to go through
// Return value of true will reject the request (ie. limit), a return value of false will pass the request.
// If Limit returns true, check the error for additional information
// Current implementations in: api_ratelimitmgr.go, full_limiter.go, flow_limiter.go, maxreqs_limiter.go, fixedwindow.go, leakybucket.go, tokenbucket.go
type Limiter interface {
	Limit(ctx context.Context) (bool, error)
}

type LimiterInfo struct {
	Api  string
	User string
	Org  string
	Ip   string
}

type limiterInfoKey struct{}

func NewLimiterInfoContext(ctx context.Context, li *LimiterInfo) context.Context {
	if li == nil {
		return ctx
	}
	return context.WithValue(ctx, limiterInfoKey{}, li)
}

func LimiterInfoFromContext(ctx context.Context) (li *LimiterInfo, ok bool) {
	li, ok = ctx.Value(limiterInfoKey{}).(*LimiterInfo)
	return
}

var DefaultReqsPerSecondPerApi = 100.0
var DefaultTokenBucketSize int64 = 10 // equivalent to burst size
