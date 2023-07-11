package v2

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	gen "simple-rpc/grpc/v2/gen"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("registry:///localhost:8080", grpc.WithInsecure(), grpc.WithResolvers(&Builder{}))
	require.NoError(t, err)

	userServiceClient := gen.NewUserServiceClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := userServiceClient.GetById(ctx, &gen.GetByIdReq{Id: 1})
	require.NoError(t, err)

	t.Log(resp)
}
