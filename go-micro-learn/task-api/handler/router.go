package handler

import (
	"github.com/gin-gonic/gin"
	"go-micro-learn/common/util/web"
	pb "go-micro-learn/task-srv/proto/task"
	"log"
	"net/http"
)

var service pb.TaskService

func Router(taskService pb.TaskService) *gin.Engine{
	service = taskService
	ginRouter := gin.Default()
	ginRouter.POST("/users", func(context *gin.Context) {
		context.JSON(http.StatusOK,gin.H{
			"code": 200,
			"msg": "请求成功",
		})
	})
	v1 := ginRouter.Group("/task")
	{
		v1.GET("/search", Search)
		v1.POST("/finished", Finished)
	}

	return ginRouter
}

func Search(c *gin.Context) {
	req := new(pb.SearchRequest)
	if err := c.BindQuery(req); err != nil {
		log.Print("bad request param: ", err)
		return
	}
	if resp, err := service.Search(c, req); err != nil {
		c.JSON(http.StatusInternalServerError, web.Fail(err.Error()))
	} else {
		c.JSON(http.StatusOK, web.Ok(resp))
	}
}

func Finished(c *gin.Context) {
	req := new(pb.Task)
	if err := c.BindJSON(req); err != nil {
		log.Print("bad request param: ", err)
		return
	}
	if resp, err := service.Finished(c, req); err != nil {
		c.JSON(http.StatusInternalServerError, web.Fail(err.Error()))
	} else {
		c.JSON(http.StatusOK, web.Ok(resp))
	}
}
