package v3

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"simple-rpc/grpc/v3/registry"
	"time"
)

type Server struct {
	name     string
	listener net.Listener

	si       registry.ServiceInstance
	registry registry.Registry
	// 启动超时，用于启动注册中心
	timeout time.Duration

	*grpc.Server
}

func NewServer(name string, opts ...ServerOption) *Server {
	res := &Server{
		name:   name,
		Server: grpc.NewServer(),
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

type ServerOption func(server *Server)

func ServerWithRegistry(registry registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = registry
	}
}

func ServerWithTimeout(timeout time.Duration) ServerOption {
	return func(server *Server) {
		server.timeout = timeout
	}
}

func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	if s.registry != nil {
		s.si = registry.ServiceInstance{
			Name:    s.name,
			Address: listener.Addr().String(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		err = s.registry.Register(ctx, s.si)
		cancel()
		if err != nil {
			return err
		}
	}

	return s.Serve(listener)
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if s.registry != nil {
		err := s.registry.UnRegister(ctx, s.si)
		if err != nil {
			return err
		}
	}

	if s.listener != nil {
		return s.listener.Close()
	}

	s.GracefulStop()
	return nil
}
