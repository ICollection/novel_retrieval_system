package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"service/internal/config"
	"service/internal/handler"
	"service/internal/server"
	"service/internal/service"
)

func main() {
	ingest := handler.NewIngestHandler(service.NewIngestService())
	chat := handler.NewChatHandler(service.NewChatService())

	engine := gin.New()
	server.Setup(engine, config.Config.Web.ContextPath, config.Config.Web.Mode, ingest, chat)

	address := fmt.Sprintf("0.0.0.0:%d", config.Config.Web.Port)
	log.Println("服务已启动，监听地址：", address)
	if err := engine.Run(address); err != nil {
		log.Fatal("启动服务失败，详细错误信息：", err)
	}
}
