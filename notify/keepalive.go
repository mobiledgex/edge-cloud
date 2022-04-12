// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
