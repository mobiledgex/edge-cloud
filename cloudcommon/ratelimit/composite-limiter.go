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

package ratelimit

import (
	"context"
)

/*
 * Composes multiple limiters into one
 */
type CompositeLimiter struct {
	limiters []Limiter
}

func NewCompositeLimiter(limiters ...Limiter) *CompositeLimiter {
	s := CompositeLimiter{
		limiters: limiters,
	}
	return &s
}

func (c *CompositeLimiter) Limit(ctx context.Context, info *CallerInfo) error {
	for _, limiter := range c.limiters {
		err := limiter.Limit(ctx, info)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CompositeLimiter) Type() string {
	return "CompositeLimiter"
}
