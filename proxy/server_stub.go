package simple_rpc

import (
	"context"
	"fmt"
	"reflect"
	"simple-rpc/proxy/message"
	"simple-rpc/proxy/serialize"
)

type ServerStub struct {
	s           Service
	value       reflect.Value
	serializers map[uint8]serialize.Serializer
}

func (s *ServerStub) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	serializer, ok := s.serializers[req.Serializer]
	if !ok {
		return nil, fmt.Errorf("server: 未找到 serializer")
	}

	method := s.value.MethodByName(req.MethodName)

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
