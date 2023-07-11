package simple_rpc

import (
	"context"
	"errors"
	"reflect"
	"simple-rpc/proxy/message"
	"simple-rpc/proxy/serialize"
	"sync/atomic"
)

var messageId uint32 = 0

type Proxy interface {
	Invoke(ctx context.Context, req *message.Request) (*message.Response, error)
}

func InitServiceProxy(service Service, client *Client) error {
	return setFuncField(service, client, client.serializer)
}

func setFuncField(service Service, proxy Proxy, serializer serialize.Serializer) error {
	if service == nil {
		return errors.New("rpc: service 不能为 nil")
	}

	val := reflect.ValueOf(service)
	typ := reflect.TypeOf(service)
	if typ.Kind() != reflect.Pointer || val.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: 只支持指向结构体的一级指针")
	}

	val = val.Elem()
	typ = typ.Elem()

	for i := 0; i < typ.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldTyp := typ.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		fn := func(args []reflect.Value) (results []reflect.Value) {

			in := args[1].Interface()
			out := reflect.New(fieldTyp.Type.Out(0).Elem()).Interface()
			data, err := serializer.Encode(in)
			if err != nil {
				return []reflect.Value{reflect.ValueOf(out), reflect.ValueOf(err)}
			}
			req := &message.Request{
				BodyLength:  uint32(len(data)),
				ServiceName: service.Name(),
				MethodName:  fieldTyp.Name,
				MessageId:   atomic.AddUint32(&messageId, 1),
				Serializer:  serializer.Code(),
				Data:        data,
			}
			req.SetHeadLength()

			// 真正发起调用
			resp, err := proxy.Invoke(args[0].Interface().(context.Context), req)
			if err != nil {
				return []reflect.Value{reflect.ValueOf(out), reflect.ValueOf(err)}
			}

			if len(resp.Data) > 0 {
				err = serializer.Decode(resp.Data, out)
				if err != nil {
					return []reflect.Value{reflect.ValueOf(out), reflect.ValueOf(err)}
				}
			}

			retErr := reflect.Zero(reflect.TypeOf(new(error)).Elem())
			if len(resp.Error) > 0 {
				retErr = reflect.ValueOf(errors.New(string(resp.Error)))
			}

			return []reflect.Value{reflect.ValueOf(out), retErr}
		}

		fieldVal.Set(reflect.MakeFunc(fieldTyp.Type, fn))
	}

	return nil
}
