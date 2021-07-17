package main

import (
	"context"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/broker/nats"
	"github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	"github.com/pkg/errors"
	"go-micro-learn/achievement-srv/repository"
	"go-micro-learn/achievement-srv/subscriber"
	"go-micro-learn/common/tracer"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)


const (
	MongoUri = "mongodb://192.168.1.96:27018"
	ServerName = "go.micro.service.achievement"
	NatsUri    = "nats://127.0.0.1:4222"
	JaegerAddr = "127.0.0.1:6831"
)

// task-srv服务
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
		micro.Name(ServerName),
		micro.Version("latest"),
		// 配置nats为消息中间件，默认端口是4222
		micro.Broker(nats.NewBroker(
			broker.Addrs(NatsUri),
		)),
		// 配置链路追踪为jaeger
		micro.WrapSubscriber(opentracing.NewSubscriberWrapper(jaegerTracer)),
	)

	// Initialise service
	service.Init()

	// Register Handler
	handler := &subscriber.AchievementSub{
		Repo: &repository.AchievementRepoImpl{
			Conn: conn,
		},
	}
	// 这里的topic注意与task-srv注册的要一致
	if err := micro.RegisterSubscriber("go.micro.service.task.finished", service.Server(), handler.Finished); err != nil {
		log.Fatal(errors.WithMessage(err, "subscribe"))
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