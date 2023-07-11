package round_robin_weight

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/balancer"
	"sync"
	"testing"
)

func TestWeightBalancer_Pick(t *testing.T) {
	b := &Balancer{
		connections: []*WeightConnection{
			{
				sc: MockSubConn{
					name: "weight-5",
				},
				weight:          5,
				currentWeight:   5,
				efficientWeight: 5,
			},
			{
				sc: MockSubConn{
					name: "weight-4",
				},
				weight:          4,
				currentWeight:   4,
				efficientWeight: 4,
			},
			{
				sc: MockSubConn{
					name: "weight-3",
				},
				weight:          3,
				currentWeight:   3,
				efficientWeight: 3,
			},
		},
		length: 2,
		mutex:  sync.Mutex{},
	}

	res, err := b.Pick(balancer.PickInfo{})
	assert.NoError(t, err)
	assert.Equal(t, "weight-5", res.SubConn.(MockSubConn).name)

	res, err = b.Pick(balancer.PickInfo{})
	assert.NoError(t, err)
	assert.Equal(t, "weight-4", res.SubConn.(MockSubConn).name)

	res, err = b.Pick(balancer.PickInfo{})
	assert.NoError(t, err)
	assert.Equal(t, "weight-3", res.SubConn.(MockSubConn).name)
}

type MockSubConn struct {
	id   int
	name string
	balancer.SubConn
}
