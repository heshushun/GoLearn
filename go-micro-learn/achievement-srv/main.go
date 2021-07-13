package main

import (
	"context"
	"github.com/micro/go-micro/v2"
	"github.com/pkg/errors"
	"go-micro-learn/achievement-srv/repository"
	"go-micro-learn/achievement-srv/subscriber"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// 这里是我内网的mongo地址，请根据你得实际情况配置，推荐使用dockers部署
const MONGO_URI = "mongodb://192.168.1.96:27018"

// task-srv服务
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
		micro.Name("go.micro.service.achievement"),
		micro.Version("latest"),
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
	if err := micro.RegisterSubscriber("go.micro.service.task.finished", service.Server(), handler); err != nil {
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