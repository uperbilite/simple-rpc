package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"simple-rpc/grpc/v3/registry"
	"sync"
)

type Registry struct {
	c    *clientv3.Client
	sess *concurrency.Session

	cancels []context.CancelFunc
	mutex   sync.Mutex
}

func NewRegistry(c *clientv3.Client) (*Registry, error) {
	sess, err := concurrency.NewSession(c)
	if err != nil {
		return nil, err
	}
	return &Registry{
		c:    c,
		sess: sess,
	}, nil
}

func (r *Registry) Register(ctx context.Context, si registry.ServiceInstance) error {
	val, err := json.Marshal(si)
	if err != nil {
		return err
	}
	_, err = r.c.Put(ctx, r.InstanceKey(si), string(val), clientv3.WithLease(r.sess.Lease()))
	return err
}

func (r *Registry) UnRegister(ctx context.Context, si registry.ServiceInstance) error {
	_, err := r.c.Delete(ctx, r.InstanceKey(si))
	return err
}

func (r *Registry) ListInstances(ctx context.Context, serviceName string) ([]registry.ServiceInstance, error) {
	getResp, err := r.c.Get(ctx, r.ServiceKey(serviceName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	res := make([]registry.ServiceInstance, 0, len(getResp.Kvs))
	for _, kv := range getResp.Kvs {
		si := registry.ServiceInstance{}
		err = json.Unmarshal(kv.Value, &si)
		if err != nil {
			return nil, err
		}
		res = append(res, si)
	}
	return res, nil

}

func (r *Registry) Subscribe(serviceName string) (<-chan registry.Event, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r.mutex.Lock()
	r.cancels = append(r.cancels, cancel)
	r.mutex.Unlock()

	ctx = clientv3.WithRequireLeader(ctx)
	watchChan := r.c.Watch(ctx, r.ServiceKey(serviceName), clientv3.WithPrefix())

	res := make(chan registry.Event, 1)

	go func() {
		for {
			select {
			case resp := <-watchChan:
				if resp.Err() != nil {
					return
				}
				if resp.Canceled {
					return
				}
				for range resp.Events {
					res <- registry.Event{}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return res, nil
}

func (r *Registry) Close() error {
	r.mutex.Lock()
	cancels := r.cancels
	r.cancels = nil
	r.mutex.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
	return r.sess.Close()
}

func (r *Registry) ServiceKey(serviceName string) string {
	return fmt.Sprintf("/micro/%s", serviceName)
}

func (r *Registry) InstanceKey(si registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s/%s", si.Name, si.Address)
}
