package handler

import (
	"github.com/gin-gonic/gin"
	"go-micro-learn/common/util/web"
	pb "go-micro-learn/task-srv/proto/task"
	"log"
	"net/http"
)

var service pb.TaskService

func Router(g *gin.Engine, taskService pb.TaskService) {
	service = taskService
	v1 := g.Group("/task")
	{
		v1.GET("/search", Search)
		v1.POST("/finished", Finished)
	}
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
