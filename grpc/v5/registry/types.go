package registry

import (
	"context"
	"io"
)

type Registry interface {
	Register(ctx context.Context, si ServiceInstance) error
	UnRegister(ctx context.Context, si ServiceInstance) error

	ListInstances(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	Subscribe(serviceName string) (<-chan Event, error)

	io.Closer
}

type ServiceInstance struct {
	Name    string
	Address string

	// 其他和服务治理相关的信息

	Weight int64
}

type Event struct{}
