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
	services    map[string]Service
	serializers map[uint8]serialize.Serializer
}

func NewServer(addr string, opts ...ServerOption) *Server {
	res := &Server{
		addr:        addr,
		services:    make(map[string]Service, 16),
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
		server.services[service.Name()] = service
	}
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = service
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

	serializer, ok := s.serializers[req.Serializer]
	if !ok {
		return nil, fmt.Errorf("server: 未找到 serializer")
	}

	method := reflect.ValueOf(service).MethodByName(req.MethodName)

	in := reflect.New(method.Type().In(1).Elem()).Interface()
	err := serializer.Decode(req.Data, in)
	if err != nil {
		return nil, err
	}

	out := method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(in)})
	data, err := serializer.Encode(out[0].Interface())
	if err != nil {
		return nil, err
	}

	resp := &message.Response{
		BodyLength: uint32(len(data)),
		MessageId:  req.MessageId,
		Data:       data,
	}
	resp.SetHeadLength()

	return resp, nil
}
