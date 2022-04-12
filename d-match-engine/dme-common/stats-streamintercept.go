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

package dmecommon

import (
	"reflect"

	log "github.com/mobiledgex/edge-cloud/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// A wrapper for grpc.ServerStream to handle stats
type StatsStreamWrapper struct {
	grpc.ServerStream
	stats *DmeStats
	info  *grpc.StreamServerInfo
	ctx   context.Context
}

func (w *StatsStreamWrapper) Context() context.Context {
	return w.ctx
}

func (w *StatsStreamWrapper) SendMsg(m interface{}) error {
	log.SpanLog(w.Context(), log.DebugLevelDmereq, "SendMsg stats stream interceptor", "type", reflect.TypeOf(m).String())
	return w.ServerStream.SendMsg(m)
}

func (w *StatsStreamWrapper) RecvMsg(m interface{}) error {
	log.SpanLog(w.Context(), log.DebugLevelDmereq, "RecvMsg stats stream interceptor", "type", reflect.TypeOf(m).String())
	return w.ServerStream.RecvMsg(m)
}

func (s *DmeStats) GetStreamStatsInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Set up a child span for the stream interceptor
		span := log.StartSpan(log.DebugLevelDmereq, "stream stats interceptor")
		defer span.Finish()
		cctx := log.ContextWithSpan(ss.Context(), span)

		wrapper := &StatsStreamWrapper{ServerStream: ss, stats: s, info: info, ctx: cctx}
		return handler(srv, wrapper)
	}
}
