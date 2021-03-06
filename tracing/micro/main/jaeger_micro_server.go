package main

import (
	jaeger_micro "GoLearn/tracing/micro"
	"GoLearn/tracing/micro/proto"
	"context"
	"fmt"
	"github.com/micro/go-micro/v2"
	wrapperTrace "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	"github.com/opentracing/opentracing-go"
	"log"
)


func main() {
	_, closer, err := jaeger_micro.NewJaegerTracer("jaeger-micro-server")
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	service := micro.NewService(
		micro.Name("jaeger.micro.server"),
		micro.Version("v2"),
		micro.WrapHandler(wrapperTrace.NewHandlerWrapper(opentracing.GlobalTracer())),
	)
	service.Init()

	_ = proto.RegisterHelloHandler(service.Server(), new(SayHello))

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

type SayHello struct{}

func (h *SayHello) Hello(ctx context.Context, req *proto.HelloRequest, rsp *proto.HelloResponse) error {
	rsp.Greeting = "Hello " + req.Name
	return nil
}
