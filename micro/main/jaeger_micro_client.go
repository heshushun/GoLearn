package main

import (
	jaeger_micro "GoLearn/micro"
	"GoLearn/micro/proto"
	"context"
	"fmt"
	"github.com/micro/go-micro/v2"
	wrapperTrace "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	"github.com/opentracing/opentracing-go"
	"log"
	"time"
)

func main(){
	tracer, closer, err := jaeger_micro.NewJaegerTracer("jaeger-micro-client")
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()
	span := tracer.StartSpan("span-micro-client")
	ctx := opentracing.ContextWithSpan(context.Background(), span)

	// 创建一个 micro Service
	service := micro.NewService(
		micro.Name("jaeger.micro.client"),
		micro.Version("v2"),
		micro.WrapClient(wrapperTrace.NewClientWrapper(opentracing.GlobalTracer())),
	)
	service.Init()

	ticker := time.NewTicker(1 * time.Second)
	for t := range ticker.C {
		serv := proto.NewHelloService("jaeger.micro.server", service.Client())
		rsp, err := serv.Hello(ctx, &proto.HelloRequest{Name: t.Format("15:04:05")})
		if err == nil {
			fmt.Printf("%v: receive is %s\n", t.Format("2006-01-02 15:04:05"), rsp.Greeting)
		} else {
			fmt.Println(err)
		}
	}

	span.Finish()
}
