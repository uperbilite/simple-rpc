package round_robin

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/balancer"
	"testing"
)

func TestBalancer_Pick(t *testing.T) {
	testCases := []struct {
		name string
		b    *Balancer

		wantSubConn MockSubConn
		wantIndex   int32
		wantDone    func(info balancer.DoneInfo)
		wantErr     error
	}{
		{
			name: "start",
			b: &Balancer{
				connections: []balancer.SubConn{
					MockSubConn{id: 1},
					MockSubConn{id: 2},
					MockSubConn{id: 3},
				},
				length: 3,
				index:  -1,
			},
			wantSubConn: MockSubConn{
				id: 1,
			},
			wantIndex: 0,
			wantDone: func(info balancer.DoneInfo) {

			},
		},
		{
			name: "end",
			b: &Balancer{
				connections: []balancer.SubConn{
					MockSubConn{id: 1},
					MockSubConn{id: 2},
					MockSubConn{id: 3},
				},
				length: 3,
				index:  1,
			},
			wantSubConn: MockSubConn{
				id: 3,
			},
			wantIndex: 2,
			wantDone: func(info balancer.DoneInfo) {

			},
		},
		{
			name: "round",
			b: &Balancer{
				connections: []balancer.SubConn{
					MockSubConn{id: 1},
					MockSubConn{id: 2},
					MockSubConn{id: 3},
				},
				length: 3,
				index:  2,
			},
			wantSubConn: MockSubConn{
				id: 1,
			},
			wantIndex: 3,
			wantDone: func(info balancer.DoneInfo) {

			},
		},
		{
			name: "no sub conn",
			b: &Balancer{
				connections: []balancer.SubConn{},
				length:      0,
				index:       -1,
			},
			wantErr: balancer.ErrNoSubConnAvailable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.b.Pick(balancer.PickInfo{})
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantIndex, tc.b.index)
			assert.Equal(t, tc.wantSubConn, res.SubConn.(MockSubConn))
			assert.NotNil(t, tc.wantDone)
		})
	}
}

type MockSubConn struct {
	id   int
	name string
	balancer.SubConn
}
