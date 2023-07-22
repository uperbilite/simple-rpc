package simple_rpc

import (
	"context"
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
