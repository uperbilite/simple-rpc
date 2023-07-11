package v5

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"simple-rpc/grpc/v5/gen"
	"simple-rpc/grpc/v5/registry/etcd"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	us := &UserServiceServer{}
	server := NewServer("user-service", ServerWithRegistry(r), ServerWithTimeout(3*time.Second))
	gen.RegisterUserServiceServer(server, us)

	err = server.Start(":8080")
	t.Log(err)
}

type UserServiceServer struct {
	gen.UnimplementedUserServiceServer
}

func (s UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id: 2,
		},
	}, nil
}
