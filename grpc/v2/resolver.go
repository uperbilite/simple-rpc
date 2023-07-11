package v2

import (
	"google.golang.org/grpc/resolver"
)

type Builder struct{}

func (b Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	res := Resolver{
		cc: cc,
	}
	res.ResolveNow(resolver.ResolveNowOptions{})
	return res, nil
}

func (b Builder) Scheme() string {
	return "registry"
}

type Resolver struct {
	cc resolver.ClientConn
}

func (r Resolver) ResolveNow(options resolver.ResolveNowOptions) {
	err := r.cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{
			{
				Addr: "localhost:8080",
			},
		},
	})
	if err != nil {
		r.cc.ReportError(err)
	}
}

func (r Resolver) Close() {
	//TODO implement me
	panic("implement me")
}
