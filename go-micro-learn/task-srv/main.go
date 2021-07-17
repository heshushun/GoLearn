package main

import (
	"context"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/broker/nats"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	"github.com/pkg/errors"
	"go-micro-learn/common/tracer"
	"go-micro-learn/task-srv/handler"
	pb "go-micro-learn/task-srv/proto/task"
	"go-micro-learn/task-srv/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

const (
	MongoUri = "mongodb://192.168.1.96:27018"
	ServerName = "go.micro.service.task"
	NatsUri    = "nats://127.0.0.1:4222"
	EtcdAddr   = "127.0.0.1:2379"
	JaegerAddr = "127.0.0.1:6831"
)

func main() {
	// 在日志中打印文件路径，便于调试代码
	log.SetFlags(log.Llongfile)

	conn, err := connectMongo(MongoUri, time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Disconnect(context.Background())

	// 配置jaeger连接
	jaegerTracer, closer, err := tracer.NewJaegerTracer(ServerName, JaegerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	// New Service
	service := micro.NewService(
		micro.Name("go.micro.service.task"),
		micro.Version("latest"),
		// 配置etcd为注册中心，配置etcd路径，默认端口是2379
		micro.Registry(etcd.NewRegistry(
			registry.Addrs(EtcdAddr),
		)),
		// 配置nats为消息中间件，默认端口是4222
		micro.Broker(nats.NewBroker(
			broker.Addrs(NatsUri),
		)),
		// 配置链路追踪为jaeger
		micro.WrapHandler(opentracing.NewHandlerWrapper(jaegerTracer)),
	)

	// Initialise service
	service.Init()

	// Register Handler
	taskHandler := &handler.TaskHandler{
		TaskRepository: &repository.TaskRepositoryImpl{
			Conn: conn,
		},
		// 注入消息发送实例,为避免消息名冲突,这里的topic我们用服务名+自定义消息名拼出
		TaskFinishedPubEvent: micro.NewEvent(handler.TaskFinishedTopic, service.Client()),
	}
	if err := pb.RegisterTaskServiceHandler(service.Server(), taskHandler); err != nil {
		log.Fatal(errors.WithMessage(err, "register server"))
	}

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(errors.WithMessage(err, "run server"))
	}
}

// 连接到MongoDB
func connectMongo(uri string, timeout time.Duration) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.WithMessage(err, "create mongo connection session")
	}
	return client, nil
}