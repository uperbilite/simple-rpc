package simple_rpc

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"simple-rpc/proxy/message"
	"simple-rpc/proxy/serialize/json"
	"testing"
)

func Test_setFuncField(t *testing.T) {
	serializer := json.Serializer{}

	testCases := []struct {
		name     string
		service  *mockService
		proxy    *mockProxy
		wantResp any
		wantErr  error
	}{
		{
			name: "user service",
			service: func() *mockService {
				s := &UserServiceClient{}
				return &mockService{
					s: s,
					do: func() (*AnyResponse, error) {
						return s.GetById(context.Background(), &AnyRequest{Msg: "这是GetById"})
					},
				}
			}(),
			proxy: &mockProxy{
				t: t,
				req: &message.Request{
					HeadLength:  36,
					BodyLength:  23,
					MessageId:   1,
					ServiceName: "user-service",
					MethodName:  "GetById",
					Serializer:  serializer.Code(),
					Data:        []byte(`{"msg":"这是GetById"}`),
				},
				resp: &message.Response{
					Data: []byte(`{"msg":"这是GetById的响应"}`),
				},
			},
			wantResp: &AnyResponse{
				Msg: "这是GetById的响应",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := setFuncField(tc.service.s, tc.proxy, &json.Serializer{})
			require.NoError(t, err)
			resp, err := tc.service.do()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantResp, resp)
		})
	}

}

type mockService struct {
	s  Service
	do func() (*AnyResponse, error)
}

type UserServiceClient struct {
	GetById func(ctx context.Context, req *AnyRequest) (*AnyResponse, error)
}

func (u *UserServiceClient) Name() string {
	return "user-service"
}

type AnyRequest struct {
	Msg string `json:"msg"`
}

type AnyResponse struct {
	Msg string `json:"msg"`
}

type mockProxy struct {
	t    *testing.T
	req  *message.Request
	resp *message.Response
	err  error
}

func (m *mockProxy) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	assert.Equal(m.t, m.req, req)
	return m.resp, m.err
}

func TestInitClientProxy(t *testing.T) {
	server := NewServer(":8080")
	server.RegisterService(&mockServiceServer{})
	go func() {
		server.StartAndServe()
	}()

	client := NewClient(":8080")
	service := &mockServiceClient{}

	err := InitServiceProxy(service, client)
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
