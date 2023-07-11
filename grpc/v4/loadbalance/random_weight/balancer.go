package random_weight

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
)

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	var totalWeight int64
	length := len(info.ReadySCs)
	connections := make([]*WeightConnection, 0, length)

	for subConn, subConnInfo := range info.ReadySCs {
		weight := subConnInfo.Address.Attributes.Value("weight").(int64)
		totalWeight += weight
		connections = append(connections, &WeightConnection{
			weight: weight,
			sc:     subConn,
		})
	}

	return &Balancer{
		connections: connections,
		length:      length,
		totalWeight: totalWeight,
	}
}

type Balancer struct {
	connections []*WeightConnection
	length      int
	totalWeight int64
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.length == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	target := rand.Int63n(b.totalWeight + 1)
	var resIdx int
	for i, c := range b.connections {
		target -= c.weight
		if target <= 0 {
			resIdx = i
			break
		}
	}
	return balancer.PickResult{
		SubConn: b.connections[resIdx].sc,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type WeightConnection struct {
	weight int64
	sc     balancer.SubConn
}
