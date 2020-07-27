// MIT License
// Copyright (c) 2020 ysicing <i@ysicing.me>

package main

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"k8s.io/klog"
	"time"
)

type ServiceReg struct {
	client *clientv3.Client
	lease clientv3.Lease
	leaseResp *clientv3.LeaseGrantResponse
	canclefunc func()
	keepaliveChan <- chan *clientv3.LeaseKeepAliveResponse
	key string
}

func NewServiceReg(addr []string) (*ServiceReg, error) {
	conf := clientv3.Config{
		Endpoints:            addr,
		DialTimeout:          time.Second * 3,
	}
	var client *clientv3.Client
	if clientTem, err := clientv3.New(conf); err == nil {
		client = clientTem
	} else {
		return nil, err
	}
	ser := &ServiceReg{
		client:        client,
	}
	if err := ser.setLease(); err != nil {
		return nil, err
	}
	return ser, nil
}

func (s *ServiceReg) setLease() error {
	lease := clientv3.NewLease(s.client)
	// 设置租约时间
	leaseResp, err := lease.Grant(context.TODO(), 5)
	if err != nil {
		return err
	}
	// 设置续租
	ctx, cancelFunc := context.WithCancel(context.TODO())
	leaseRespChan, err := lease.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}
	s.lease = lease
	s.leaseResp = leaseResp
	s.canclefunc = cancelFunc
	s.keepaliveChan = leaseRespChan
	return nil
}

func (s *ServiceReg) ListenLeaseRespChan()  {
	for {
		select {
		case leaseKeepResp := <- s.keepaliveChan:
			if leaseKeepResp == nil {
				return
			}

		}
	}
}

func (s *ServiceReg) PutService(key, val string) error  {
	kv := clientv3.NewKV(s.client)
	_, err := kv.Put(context.TODO(), key, val, clientv3.WithLease(s.leaseResp.ID))
	return err
}

func (s *ServiceReg) RevokeLease() error  {
	s.canclefunc()
	time.Sleep(time.Second * 2)
	_, err := s.lease.Revoke(context.TODO(), s.leaseResp.ID)
	return err
}

func main()  {
	svc, err  := NewServiceReg([]string{"127.0.0.1:2379"})
	if err != nil {
		klog.Error(err)
	}
	if err = svc.PutService("/node/127.0.0.1", "localhost"); err != nil {
		klog.Error(err)
	}
}