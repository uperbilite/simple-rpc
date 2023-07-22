package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type FixedWindowLimiter struct {
	// 窗口起始时间
	timestamp int64
	// 窗口大小
	interval int64
	// 允许通过的最大数量
	rate  int64
	cnt   int64
	mutex sync.Mutex
}

func NewFixedWindowLimiter(interval time.Duration, rate int64) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		timestamp: time.Now().UnixNano(),
		interval:  interval.Nanoseconds(),
		rate:      rate,
	}
}

func (f *FixedWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		f.mutex.Lock()
		current := time.Now().UnixNano()
		if f.timestamp+f.interval < current {
			f.timestamp = current
			f.cnt = 0
		}
		if f.cnt >= f.rate {
			f.mutex.Unlock()
			err = errors.New("触发瓶颈")
			return
		}
		f.cnt++
		f.mutex.Unlock()
		resp, err = handler(ctx, req)
		return
	}
}
