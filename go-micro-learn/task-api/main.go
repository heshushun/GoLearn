package main

import (
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/web"
	"github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	"go-micro-learn/common/tracer"
	"go-micro-learn/task-api/handler"
	"go-micro-learn/task-api/wrapper/breaker/hystrix"
	pb "go-micro-learn/task-srv/proto/task"
	"log"
)

const (
	ServerName = "go.micro.api.task"
	EtcdAddr   = "127.0.0.1:2379"
	JaegerAddr = "127.0.0.1:6831"
)

// task-srv服务的restful api映射
func main() {

	// 配置jaeger连接
	jaegerTracer, closer, err := tracer.NewJaegerTracer(ServerName, JaegerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	etcdRegister := etcd.NewRegistry(
		registry.Addrs(EtcdAddr),
	)
	// 之前我们使用client.DefaultClient注入到pb.NewTaskService中
	// 现在改为标准的服务创建方式创建服务对象
	// 但这个服务并不真的运行（我们并不调用他的Init()和Run()方法）
	// 如果是task-srv这类本来就是用micro.NewService服务创建的服务
	// 则直接增加包装器，不需要再额外新增服务
	app := micro.NewService(
		micro.Name("go.micro.client.task"),
		micro.Registry(etcdRegister),
		micro.WrapClient(
			// 引入hystrix包装器
			hystrix.NewClientWrapper(),
			// 配置链路追踪为jaeger
			opentracing.NewClientWrapper(jaegerTracer),
		),
	)
	taskService := pb.NewTaskService("go.micro.service.task", app.Client())

	// 配置web路由
	g := handler.Router(taskService)
	// 这个服务才是真正运行的服务
	service := web.NewService(
		web.Name("go.micro.api.task"),
		web.Address(":8888"),
		web.Handler(g),
		web.Registry(etcdRegister),
	)

	_ = service.Init()
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
