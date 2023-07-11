package round_robin_weight

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync"
)

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	length := len(info.ReadySCs)
	connections := make([]*WeightConnection, 0, length)

	for subConn, subConnInfo := range info.ReadySCs {
		weight := subConnInfo.Address.Attributes.Value("weight").(int64)
		connections = append(connections, &WeightConnection{
			weight:          weight,
			currentWeight:   weight,
			efficientWeight: weight,
			sc:              subConn,
		})
	}

	return &Balancer{
		connections: connections,
		length:      length,
	}
}

type Balancer struct {
	connections []*WeightConnection
	length      int
	mutex       sync.Mutex
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	var totalWeight int64
	var res *WeightConnection
	for _, wc := range b.connections {
		totalWeight += wc.efficientWeight
		wc.currentWeight = wc.currentWeight + wc.efficientWeight
		if res == nil || res.currentWeight < wc.currentWeight {
			res = wc
		}
	}
	res.currentWeight -= totalWeight

	return balancer.PickResult{
		SubConn: res.sc,
		Done: func(info balancer.DoneInfo) {
			b.mutex.Lock()
			defer b.mutex.Unlock()
			if info.Err != nil && res.efficientWeight == 0 {
				return
			}
			if info.Err == nil && res.efficientWeight == math.MaxInt64 {
				return
			}
			if info.Err != nil {
				res.efficientWeight--
			} else {
				res.efficientWeight++
			}
		},
	}, nil
}

type WeightConnection struct {
	weight          int64
	currentWeight   int64
	efficientWeight int64
	sc              balancer.SubConn
}
