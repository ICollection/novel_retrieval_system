package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"service/internal/config"
	"service/internal/server"
)

func main() {

	engine := gin.New()
	server.Setup(engine, config.Config.Web.ContextPath, config.Config.Web.Mode, nil, nil)
	address := fmt.Sprintf("0.0.0.0:%d", config.Config.Web.Port)
	log.Println("服务已启动，监听地址：", address)
	if err := engine.Run(address); err != nil {
		log.Fatal("启动服务失败，详细错误信息：", err)
	}
}
