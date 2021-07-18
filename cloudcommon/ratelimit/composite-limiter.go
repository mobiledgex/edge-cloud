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
