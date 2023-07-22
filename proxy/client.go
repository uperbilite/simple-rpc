package simple_rpc

import (
	"context"
	"github.com/silenceper/pool"
	"net"
	"simple-rpc/proxy/message"
	"simple-rpc/proxy/serialize"
	"simple-rpc/proxy/serialize/json"
	"time"
)

type Client struct {
	addr       string
	connPool   pool.Pool
	serializer serialize.Serializer
}

func NewClient(addr string, opts ...ClientOption) *Client {
	poolConfig := &pool.Config{
		InitialCap:  5,
		MaxIdle:     20,
		MaxCap:      30,
		Factory:     func() (interface{}, error) { return net.Dial("tcp", addr) },
		Close:       func(v interface{}) error { return v.(net.Conn).Close() },
		IdleTimeout: time.Minute,
	}
	connPool, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil
	}
	res := &Client{
		addr:       addr,
		serializer: &json.Serializer{},
		connPool:   connPool,
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
	conn, err := c.connPool.Get()
	defer c.connPool.Put(conn)
	if err != nil {
		return nil, err
	}

	data := message.EncodeReq(req)
	Write(conn.(net.Conn), data)

	data = Read(conn.(net.Conn))
	res := message.DecodeResp(data)

	return res, nil
}
