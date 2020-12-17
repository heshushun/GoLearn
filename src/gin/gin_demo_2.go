package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func main()  {
	route := gin.New()

	// LoggerWithFormatter 中间件会将日志写入 gin.DefaultWriter
	// 自定义日志格式
	route.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	route.Use(gin.Recovery())
	// 添加路由
	route.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// 监听、启动服务
	route.Run(":8080")

}