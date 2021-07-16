package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/pkg/errors"
	pb "go-micro-learn/task-srv/proto/task"
	"go-micro-learn/task-srv/repository"
	"log"
	"net/http"
	"time"
)

var etcdRegC registry.Registry

func init(){
	etcdRegC = etcd.NewRegistry(
		registry.Addrs("127.0.0.1:2379"),
	)
}

func main() {
	// 在日志中打印文件路径，便于调试代码
	log.SetFlags(log.Llongfile)
	// 客户端也注册为服务
	server := micro.NewService(
		micro.Name("go.micro.client.task"),
		// 配置etcd为注册中心，配置etcd路径，默认端口是2379
		micro.Registry(etcdRegC),
	)
	server.Init()

	// etcd 获取服务地址
	hostAddress := getServiceAddr("go.micro.api.task")
	if len(hostAddress) <= 0 {
		fmt.Println("hostAddress is null")
	}else{
		url := "http://" + hostAddress + "/users"
		log.Println(url)
		resp, _ := http.Post(url,"application/json;charset=utf-8",bytes.NewBuffer([]byte("")))
		log.Println(resp)
	}

	taskService := pb.NewTaskService("go.micro.service.task", server.Client())
	doTaskService(taskService)

	// Run client
	if err := server.Run(); err != nil {
		log.Fatal(errors.WithMessage(err, "run client"))
	}

}

func insertTask(taskService pb.TaskService, body string, start, end int64) {
	_, err := taskService.Create(context.Background(), &pb.Task{
		Body:      body,
		StartTime: start,
		EndTime:   end,
		// 这里先随便输入一个userId
		UserId: "10000",
	})
	if err != nil {
		log.Fatal("create", err)
	}
	log.Println("create task success! ")
}

// 执行task操作
func doTaskService(taskService pb.TaskService){

	// 1.调用服务生成三条任务
	now := time.Now()
	//insertTask(taskService, "完成学习笔记（一）", now.Unix(), now.Add(time.Hour*24).Unix())
	//insertTask(taskService, "完成学习笔记（二）", now.Add(time.Hour*24).Unix(), now.Add(time.Hour*48).Unix())
	//insertTask(taskService, "完成学习笔记（三）", now.Add(time.Hour*48).Unix(), now.Add(time.Hour*72).Unix())

	// 2.分页查询任务列表
	page, err := taskService.Search(context.Background(), &pb.SearchRequest{
		PageCode: 1,
		PageSize: 20,
	})
	if err != nil {
		log.Fatal("search1", err)
	}
	log.Println(page)

	// 3.更新第一条记录为完成
	row := page.Rows[0]
	if _, err = taskService.Finished(context.Background(), &pb.Task{
		Id:         row.Id,
		IsFinished: repository.Finished,
	}); err != nil {
		log.Fatal("finished", row.Id, err)
	}

	// 4.修改查询到的第二条数据,延长截至日期
	row = page.Rows[1]
	if _, err = taskService.Modify(context.Background(), &pb.Task{
		Id:        row.Id,
		Body:      row.Body,
		StartTime: row.StartTime,
		EndTime:   now.Add(time.Hour * 72).Unix(),
	}); err != nil {
		log.Fatal("modify", row.Id, err)
	}

	// 5.删除第三条记录
	//row = page.Rows[2]
	//if _, err = taskService.Delete(context.Background(), &pb.Task{
	//	Id: row.Id,
	//}); err != nil {
	//	log.Fatal("delete", row.Id, err)
	//}

	// 6.再次分页查询，校验修改结果
	page, err = taskService.Search(context.Background(), &pb.SearchRequest{})
	if err != nil {
		log.Fatal("search2", err)
	}
	log.Println(page)
}

// etcd 获取服务地址
func getServiceAddr(serviceName string)(address string){
	var retryCount  int
	for {
		servers,err := etcdRegC.GetService(serviceName)
		if err != nil{
			fmt.Println(err.Error())
		}
		var services []*registry.Service
		for _,value := range servers{
			fmt.Println(value.Name, ":", value.Version)
			services = append(services, value)
		}
		next := selector.RoundRobin(services)
		if node, err := next();err == nil{
			address = node.Address
		}
		if len(address) > 0 {
			return
		}
		// 重试次数
		retryCount  ++
		time.Sleep(time.Second * 3)
		if retryCount  >= 5{
			return
		}
	}
}

