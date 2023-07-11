package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	length := len(info.ReadySCs)
	connections := make([]balancer.SubConn, 0, length)
	for c, _ := range info.ReadySCs {
		connections = append(connections, c)
	}
	return &Balancer{
		connections: connections,
		length:      length,
		index:       -1,
	}
}

type Balancer struct {
	connections []balancer.SubConn
	length      int
	index       int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := atomic.AddInt32(&b.index, 1)
	c := b.connections[idx%int32(b.length)]
	return balancer.PickResult{
		SubConn: c,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}
