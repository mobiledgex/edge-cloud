package notify

import (
	"math"
	"time"

	"google.golang.org/grpc/keepalive"
)

// Keepalive parameters to close the connection if the other end
// goes away unexpectedly. The server and client parameters must be balanced
// correctly or the connection may be closed incorrectly.
const (
	infinity   = time.Duration(math.MaxInt64)
	kpInterval = 30 * time.Second
)

var serverParams = keepalive.ServerParameters{
	MaxConnectionIdle:     3 * kpInterval,
	MaxConnectionAge:      infinity,
	MaxConnectionAgeGrace: infinity,
	Time:    kpInterval,
	Timeout: kpInterval,
}
var clientParams = keepalive.ClientParameters{
	Time:    kpInterval,
	Timeout: kpInterval,
}
var serverEnforcement = keepalive.EnforcementPolicy{
	MinTime: 1 * time.Second,
}
