package simple_rpc

import (
	"context"
	"net"
	"simple-rpc/proxy/message"
	"simple-rpc/proxy/serialize"
	"simple-rpc/proxy/serialize/json"
)

type Client struct {
	addr       string
	serializer serialize.Serializer
}

func NewClient(addr string, opts ...ClientOption) *Client {
	res := &Client{
		addr:       addr,
		serializer: &json.Serializer{},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

type ClientOption func(client *Client)

func ClientWithSerializer(serializer serialize.Serializer) ClientOption {
	return func(client *Client) {
		client.serializer = serializer
	}
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return nil, err
	}

	data := message.EncodeReq(req)
	Write(conn, data)

	//if isOneway(ctx) {
	//	return &message.Response{}, errors.New("client: 这是 oneway 调用")
	//}

	data = Read(conn)
	res := message.DecodeResp(data)

	return res, nil
}
