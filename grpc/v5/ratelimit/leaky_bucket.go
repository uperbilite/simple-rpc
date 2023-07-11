package ratelimit

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type LeakyBucketLimiter struct {
	ticker *time.Ticker
}

func NewLeakyBucketLimiter(interval time.Duration) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		ticker: time.NewTicker(interval),
	}
}

func (l *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-l.ticker.C:
			return handler(ctx, req)
		}
	}
}

func (l *LeakyBucketLimiter) Close() error {
	l.ticker.Stop()
	return nil
}
