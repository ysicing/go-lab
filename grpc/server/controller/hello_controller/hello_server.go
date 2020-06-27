// MIT License
// Copyright (c) 2020 ysicing <i@ysicing.me>

package hello_controller

import (
	"context"
	"fmt"
	"github.com/ysicing/go-lab/grpc/server/proto/hello"
	"time"
)

type HelloController struct {}

func (h *HelloController) SayHello(ctx context.Context, in *hello.HelloRequest) (*hello.HelloResponse, error) {
	return &hello.HelloResponse{
		Name: fmt.Sprintf("%s", in.Name)}, nil
}

func (h *HelloController) LotsOfReplies(in *hello.HelloRequest, steam hello.Hello_LotsOfRepliesServer) error {
	for i := 0; i <= 10; i ++ {
		steam.Send(&hello.HelloResponse{
			Name: fmt.Sprintf("%s reply %d on %v", in.Name, i, time.Now().UnixNano()),
		})
	}
	return nil
}