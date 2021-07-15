package main

import (
	"context"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/pkg/errors"
	"go-micro-learn/task-srv/handler"
	pb "go-micro-learn/task-srv/proto/task"
	"go-micro-learn/task-srv/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// 这里是我内网的mongo地址，请根据你得实际情况配置，推荐使用dockers部署
const MONGO_URI = "mongodb://192.168.1.96:27018"

func main() {
	// 在日志中打印文件路径，便于调试代码
	log.SetFlags(log.Llongfile)

	conn, err := connectMongo(MONGO_URI, time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Disconnect(context.Background())

	// New Service
	service := micro.NewService(
		micro.Name("go.micro.service.task"),
		micro.Version("latest"),
		// 配置etcd为注册中心，配置etcd路径，默认端口是2379
		micro.Registry(etcd.NewRegistry(
			registry.Addrs("127.0.0.1:2379"),
		)),
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