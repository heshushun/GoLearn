package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"os"
)

func main() {
	// 禁用控制台颜色
	gin.DisableConsoleColor()

	// 创建记录日志的文件
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	// 如果需要将日志同时写入文件和控制台，请使用以下代码
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	// 初始化创建 engine
	r := gin.Default()
	// 添加路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// 监听、启动服务
	r.Run(":8080")
}


