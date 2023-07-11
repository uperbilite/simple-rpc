package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"time"
)

type TokenBucketLimiter struct {
	tokens  chan struct{}
	closeCh chan struct{}
}

func NewTokenBucketLimiter(capacity int, interval time.Duration) *TokenBucketLimiter {
	tokens := make(chan struct{}, capacity)
	closeCh := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case tokens <- struct{}{}:
				default:
					// 防止没人取令牌导致tokens阻塞，导致无法被关闭，因为此时不能进入closeCh的case
				}
			case <-closeCh:
				return
				// 这里不加default防止循环空转
			}
		}
	}()

	return &TokenBucketLimiter{
		tokens:  tokens,
		closeCh: closeCh,
	}
}

func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-t.closeCh:
			// 已经关闭故障检测，不再进行限流
			// return handler(ctx, req)
			err = errors.New("close")
			return
		case <-t.tokens:
			return handler(ctx, req)
		}
	}
}

func (t *TokenBucketLimiter) Close() error {
	close(t.closeCh)
	return nil
}
