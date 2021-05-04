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
	Limit(ctx Context) (bool, error)
}

type Context struct {
	context.Context
	Api  string
	User string
	Org  string
	Ip   string
}
