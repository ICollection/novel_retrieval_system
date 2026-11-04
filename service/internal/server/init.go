package server

import (
	"fmt"
	"log"
	"service/internal/config"

	"github.com/gin-gonic/gin"
)

// Engine 服务引擎。
var Engine *gin.Engine = nil

// init 初始化服务器。
func init() {
	// 设置运行模式。
	switch config.Config.Web.Mode {
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
	Engine := gin.Default()
	Engine.Use(CrossDomain)
	// 启动网络服务。
	address := fmt.Sprintf("0.0.0.0:%d", config.Config.Web.Port)
	log.Println("正在启动服务，服务监听地址：", address)
	if err := Engine.Run(address); err != nil {
		log.Fatal("启动服务失败，详细错误信息：", err)
	}
}
