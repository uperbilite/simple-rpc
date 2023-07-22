package simple_rpc

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitClientProxy(t *testing.T) {
	server := NewServer(":8080")
	server.RegisterService(&mockServiceServer{})
	go func() {
		server.StartAndServe()
	}()

	client := NewClient(":8080")
	service := &mockServiceClient{}

	err := InitClientProxy(service, client)
	assert.NoError(t, err)

	testCases := []struct {
		name     string
		id       int
		wantResp string
		wantErr  error
	}{
		{
			name:     "normal",
			id:       1,
			wantResp: fmt.Sprintf("hello, id: %d", 1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := service.MockFunc(context.Background(), &MockReqClient{Id: tc.id})
			if err != nil {
				assert.Equal(t, tc.wantErr, err)
			}
			assert.Equal(t, tc.wantResp, resp.Resp)
		})
	}

}

type mockServiceClient struct {
	MockFunc func(ctx context.Context, req *MockReqClient) (*MockRespClient, error)
}

type MockReqClient struct {
	Id int
}

type MockRespClient struct {
	Resp string
}

func (m *mockServiceClient) Name() string {
	return "mock-service"
}

type mockServiceServer struct {
}

type MockReqServer struct {
	Id int
}

type MockRespServer struct {
	Resp string
}

func (m *mockServiceServer) Name() string {
	return "mock-service"
}

func (m *mockServiceServer) MockFunc(ctx context.Context, req *MockReqServer) (*MockRespServer, error) {
	return &MockRespServer{
		Resp: fmt.Sprintf("hello, id: %d", req.Id),
	}, nil
}
