package ratelimit

import (
	"container/list"
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	queue    *list.List
	interval int64
	rate     int
	mutex    sync.Mutex
}

func NewSlideWindowLimiter(interval time.Duration, rate int) *SlideWindowLimiter {
	return &SlideWindowLimiter{
		queue:    new(list.List),
		interval: interval.Nanoseconds(),
		rate:     rate,
	}
}

func (s *SlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		current := time.Now().UnixNano()
		s.mutex.Lock()

		// 快路径
		if s.queue.Len() < s.rate {
			s.mutex.Unlock()
			resp, err = handler(ctx, req)
			s.queue.PushBack(current)
			return
		}

		// 慢路径
		timestamp := s.queue.Front()
		for timestamp != nil && timestamp.Value.(int64) < current-s.interval {
			s.queue.Remove(timestamp)
			timestamp = s.queue.Front()
		}
		if s.queue.Len() >= s.rate {
			err = errors.New("触发瓶颈")
			s.mutex.Unlock()
			return
		}

		s.mutex.Unlock()

		resp, err = handler(ctx, req)

		s.queue.PushBack(current)

		return
	}
}
