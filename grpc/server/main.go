// MIT License
// Copyright (c) 2020 ysicing <i@ysicing.me>

package main

import (
	"github.com/ysicing/go-lab/grpc/server/controller/hello_controller"
	"github.com/ysicing/go-lab/grpc/server/proto/hello"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	Addr = "0.0.0.0:19090"
)

func main()  {
	listen, err := net.Listen("tcp", Addr)
	if err != nil {
		log.Fatalf("listen err: %v", err)
	}
	s := grpc.NewServer()
	hello.RegisterHelloServer(s, &hello_controller.HelloController{})
	if err := s.Serve(listen); err != nil {
		log.Fatalf("serve err: %v", err)
	}
}

