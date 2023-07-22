package simple_rpc

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"simple-rpc/proxy/message"
	"simple-rpc/proxy/serialize"
	"simple-rpc/proxy/serialize/json"
)

type Server struct {
	addr        string
	services    map[string]*ServerStub
	serializers map[uint8]serialize.Serializer
}

func NewServer(addr string, opts ...ServerOption) *Server {
	res := &Server{
		addr:        addr,
		services:    make(map[string]*ServerStub, 16),
		serializers: make(map[uint8]serialize.Serializer, 4),
	}
	res.RegisterSerializer(&json.Serializer{})

	for _, opt := range opts {
		opt(res)
	}

	return res
}

type ServerOption func(server *Server)

func ServerWithSerializer(serializer serialize.Serializer) ServerOption {
	return func(server *Server) {
		server.serializers[serializer.Code()] = serializer
	}
}

func ServerWithService(service Service) ServerOption {
	return func(server *Server) {
		server.services[service.Name()] = &ServerStub{
			s:           service,
			value:       reflect.ValueOf(service),
			serializers: server.serializers,
		}
	}
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = &ServerStub{
		s:           service,
		value:       reflect.ValueOf(service),
		serializers: s.serializers,
	}
}

func (s *Server) RegisterSerializer(serializer serialize.Serializer) {
	s.serializers[serializer.Code()] = serializer
}

func (s *Server) StartAndServe() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func() {
			_ = s.handleConn(conn)
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	for {
		data := Read(conn)
		req := message.DecodeReq(data)

		resp, err := s.Invoke(context.Background(), req)
		if err != nil {
			return err
		}

		data = message.EncodeResp(resp)
		Write(conn, data)
	}
}

func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, fmt.Errorf("server: 未找到服务, 服务名 %s", req.ServiceName)
	}
	return service.Invoke(ctx, req)
}
