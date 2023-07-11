package v5

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"simple-rpc/grpc/v5/registry"
	"time"
)

type Client struct {
	registry   registry.Registry
	isInsecure bool
	// 启动超时，用于监听注册中心
	timeout time.Duration
}

func NewClient(opts ...ClientOption) (*Client, error) {
	res := &Client{}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

type ClientOption func(client *Client)

func ClientWithRegistry(registry registry.Registry) ClientOption {
	return func(client *Client) {
		client.registry = registry
	}
}

func ClientWithTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.timeout = timeout
	}
}

func ClientWithIsInsecure(isInsecure bool) ClientOption {
	return func(client *Client) {
		client.isInsecure = isInsecure
	}
}

func (c *Client) Dial(ctx context.Context, serviceName string) (*grpc.ClientConn, error) {
	address := fmt.Sprintf("registry:///%s", serviceName)
	var opts []grpc.DialOption
	if c.registry != nil {
		rb := NewResolverBuilder(c.registry, c.timeout)
		opts = append(opts, grpc.WithResolvers(rb))
	}
	if c.isInsecure {
		opts = append(opts, grpc.WithInsecure())
	}
	return grpc.DialContext(ctx, address, opts...)
}
