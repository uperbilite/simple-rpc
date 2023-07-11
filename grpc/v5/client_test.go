package v5

import (
	"context"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"simple-rpc/grpc/v5/gen"
	"simple-rpc/grpc/v5/registry/etcd"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	client, err := NewClient(ClientWithIsInsecure(true), ClientWithRegistry(r), ClientWithTimeout(3*time.Second))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cc, err := client.Dial(ctx, "user-service")
	require.NoError(t, err)

	uc := gen.NewUserServiceClient(cc)
	resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 1})
	require.NoError(t, err)

	t.Log(resp)
}
