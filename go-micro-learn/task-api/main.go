package main

import (
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/web"
	"go-micro-learn/task-api/wrapper/breaker/hystrix"
	pb "go-micro-learn/task-srv/proto/task"
	"log"
	"net/http"
)

// task-srv服务的restful api映射
func main() {
	etcdRegister := etcd.NewRegistry(
		registry.Addrs("127.0.0.1:2379"),
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
		),
	)
	taskService := pb.NewTaskService("go.micro.service.task", app.Client())

	// 配置web路由
	g := InitRouter(taskService)
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

func InitRouter(taskService pb.TaskService)*gin.Engine{
	ginRouter := gin.Default()
	ginRouter.POST("/users", func(context *gin.Context) {
		context.JSON(http.StatusOK,gin.H{
			"code": 200,
			"msg": "请求成功",
		})
	})
	v1 := ginRouter.Group("/task")
	{
		v1.GET("/search", func(c *gin.Context) {
			req := new(pb.SearchRequest)
			if err := c.BindQuery(req); err != nil {
				c.JSON(200, gin.H{
					"code": "500",
					"msg":  "bad param",
				})
				return
			}
			if resp, err := taskService.Search(c, req); err != nil {
				c.JSON(200, gin.H{
					"code": "500",
					"msg":  err.Error(),
				})
			} else {
				c.JSON(200, gin.H{
					"code": "200",
					"data": resp,
				})
			}
		})
	}
	return ginRouter
}