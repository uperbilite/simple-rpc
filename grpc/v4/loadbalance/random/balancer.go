package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
)

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	length := len(info.ReadySCs)
	connections := make([]balancer.SubConn, length)
	for connection := range info.ReadySCs {
		connections = append(connections, connection)
	}
	return &Balancer{
		connections: connections,
		length:      length,
	}
}

type Balancer struct {
	connections []balancer.SubConn
	length      int
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.length == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := rand.Intn(b.length)
	return balancer.PickResult{
		SubConn: b.connections[idx],
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}
