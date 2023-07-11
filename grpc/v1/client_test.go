package v1

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	gen "simple-rpc/grpc/v1/gen"
	"testing"
)

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	require.NoError(t, err)

	client := gen.NewUserServiceClient(cc)

	resp, err := client.GetById(context.Background(), &gen.GetByIdReq{Id: 1})
	require.NoError(t, err)

	t.Log(resp)
}
