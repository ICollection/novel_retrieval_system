package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"service/internal/config"
	"service/internal/handler"
	"service/internal/server"
	"service/internal/service"
	"service/internal/storage"
)

func main() {
	// 创建数据连接。
	postgre, err := storage.CreatePostgre()
	if err != nil {
		log.Fatal("创建 PostgreSQL 存储失败，详细错误信息：", err)
	}
	defer postgre.Close()

	// 创建对象存储。
	objects, err := storage.CreateMinio()
	if err != nil {
		log.Fatal("创建 MinIO 存储失败，详细错误信息：", err)
	}

	// 创建业务接口。
	ingest := handler.NewIngestHandler(service.NewIngestService(postgre, objects))
	chat := handler.NewChatHandler(service.NewChatService())

	engine := gin.New()
	server.Setup(engine, config.Config.Web.ContextPath, config.Config.Web.Mode, ingest, chat)
	address := fmt.Sprintf("0.0.0.0:%d", config.Config.Web.Port)
	log.Println("服务已启动，监听地址：", address)
	if err := engine.Run(address); err != nil {
		log.Fatal("启动服务失败，详细错误信息：", err)
	}
}
