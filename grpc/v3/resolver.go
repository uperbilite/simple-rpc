package v3

import (
	"context"
	"google.golang.org/grpc/resolver"
	"simple-rpc/grpc/v3/registry"
	"time"
)

type ResolverBuilder struct {
	registry registry.Registry
	timeout  time.Duration
}

func NewResolverBuilder(r registry.Registry, timeout time.Duration) ResolverBuilder {
	return ResolverBuilder{
		registry: r,
		timeout:  timeout,
	}
}

func (r ResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	res := Resolver{
		target:   target,
		cc:       cc,
		registry: r.registry,
		timeout:  r.timeout,
		close:    make(chan struct{}, 1),
	}
	res.resolve()
	go res.Watch()
	return res, nil
}

func (r ResolverBuilder) Scheme() string {
	return "registry"
}

type Resolver struct {
	target   resolver.Target
	cc       resolver.ClientConn
	registry registry.Registry
	timeout  time.Duration
	close    chan struct{}
}

func (r Resolver) ResolveNow(options resolver.ResolveNowOptions) {
	r.resolve()
}

func (r Resolver) Watch() {
	events, err := r.registry.Subscribe(r.target.Endpoint())
	if err != nil {
		return
	}
	for {
		select {
		case <-events:
			r.resolve()
		case <-r.close:
			return
		}
	}
}

func (r Resolver) Close() {
	r.close <- struct{}{}
}

func (r Resolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	instances, err := r.registry.ListInstances(ctx, r.target.Endpoint())
	cancel()
	if err != nil {
		r.cc.ReportError(err)
		return
	}

	addresses := make([]resolver.Address, 0, len(instances))
	for _, si := range instances {
		addresses = append(addresses, resolver.Address{
			ServerName: si.Name,
			Addr:       si.Address,
		})
	}

	err = r.cc.UpdateState(resolver.State{
		Addresses: addresses,
	})
	if err != nil {
		r.cc.ReportError(err)
		return
	}
}
