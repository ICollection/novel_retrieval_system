package main

import (
	"fmt"
	"log"
	"service/internal/config"
	"service/internal/handler"
	"service/internal/pipeline"
	"service/internal/server"
	"service/internal/service"
	"service/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化存储层。
	storage.Initalize()
	defer storage.Dispose()
	// 初始化服务层。
	service.Initalize()
	// 初始化处理层。
	handler.Initalize()
	// 启动任务调度。
	pipeline.Initalize()
	defer pipeline.Dispose()
	// 设置运行模式。
	switch config.Config.Station.Mode {
	case gin.ReleaseMode:
		gin.SetMode(gin.ReleaseMode)
		break
	case gin.TestMode:
		gin.SetMode(gin.TestMode)
		break
	default:
		gin.SetMode(gin.DebugMode)
		break
	}
	// 创建服务引擎。
	engine := gin.Default()
	engine.Use(server.CrossDomain)
	// 注册服务方法。
	handler.Register(engine)
	// 启动网络服务。
	address := fmt.Sprintf("0.0.0.0:%d", config.Config.Station.Port)
	log.Println("正在启动服务，服务监听地址：", address)
	if err := engine.Run(address); err != nil {
		log.Fatal("启动服务失败，详细错误信息：", err)
	}
}
