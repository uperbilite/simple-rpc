package ratelimit

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSlideWindowLimiter_BuildServerInterceptor(t *testing.T) {
	interceptor := NewSlideWindowLimiter(3*time.Second, 1).BuildServerInterceptor()
	cnt := 0
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		cnt++
		return nil, nil
	}

	resp, err := interceptor(context.Background(), nil, nil, handler)
	require.NoError(t, err)
	assert.Equal(t, nil, resp)

	resp, err = interceptor(context.Background(), nil, nil, handler)
	require.Equal(t, errors.New("触发瓶颈"), err)
	assert.Equal(t, nil, resp)

	time.Sleep(3 * time.Second)

	resp, err = interceptor(context.Background(), nil, nil, handler)
	require.NoError(t, err)
	assert.Equal(t, nil, resp)
}
