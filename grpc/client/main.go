// MIT License
// Copyright (c) 2020 ysicing <i@ysicing.me>

package main

import (
	"context"
	"github.com/ysicing/go-lab/grpc/server/proto/hello"
	"google.golang.org/grpc"
	"io"
	"k8s.io/klog"
)

const (
	Addr = "0.0.0.0:19090"
)

func main()  {
	conn, err := grpc.Dial(Addr, grpc.WithInsecure())
	if err != nil {
		klog.Errorf("conn err: %v", err)
	}
	defer conn.Close()
	c := hello.NewHelloClient(conn)
	res, err := c.SayHello(context.Background(), &hello.HelloRequest{
		Name: "aaaa"})
	if err != nil {
		klog.Error(err)
	}
	klog.Info(res)
	stream, err := c.LotsOfReplies(context.Background(), &hello.HelloRequest{Name: "aaa"})
	if err != nil {
		klog.Error(err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			klog.Errorf("stream recv: %v", err)
		}
		klog.Info(res)
	}
}
