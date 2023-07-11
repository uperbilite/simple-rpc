package ratelimit

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestTokenBucketLimiter_BuildServerInterceptor(t *testing.T) {
	testCases := []struct {
		name string

		tbl     func() *TokenBucketLimiter
		ctx     context.Context
		handler grpc.UnaryHandler

		wantResp interface{}
		wantErr  error
	}{
		{
			name: "close",
			tbl: func() *TokenBucketLimiter {
				closeCh := make(chan struct{})
				close(closeCh)
				return &TokenBucketLimiter{
					tokens:  make(chan struct{}),
					closeCh: closeCh,
				}
			},
			ctx:     context.Background(),
			wantErr: errors.New("close"),
		},
		{
			name: "cancel",
			tbl: func() *TokenBucketLimiter {
				return &TokenBucketLimiter{
					tokens:  make(chan struct{}),
					closeCh: make(chan struct{}),
				}
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
		{
			name: "token",
			tbl: func() *TokenBucketLimiter {
				tokens := make(chan struct{}, 1)
				tokens <- struct{}{}
				return &TokenBucketLimiter{
					tokens:  tokens,
					closeCh: make(chan struct{}),
				}
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return "token", nil
			},
			ctx:      context.Background(),
			wantResp: "token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interceptor := tc.tbl().BuildServerInterceptor()
			resp, err := interceptor(tc.ctx, nil, nil, tc.handler)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestTokenBucketLimiter_Tokens(t *testing.T) {
	limiter := NewTokenBucketLimiter(10, 2*time.Second)
	defer limiter.Close()
	interceptor := limiter.BuildServerInterceptor()
	cnt := 0
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		cnt++
		return nil, nil
	}
	resp, err := interceptor(context.Background(), nil, nil, handler)
	require.NoError(t, err)
	assert.Equal(t, nil, resp)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	// 触发限流
	resp, err = interceptor(ctx, nil, nil, handler)
	require.Equal(t, context.DeadlineExceeded, err)
	require.Nil(t, resp)
}
