package main

import (
	"fmt"
	"log/slog"
	"os"
	"service/internal/common"
	"service/internal/config"
	"service/internal/handler"
	"service/internal/pipeline"
	"service/internal/server"
	"service/internal/service"
	"service/internal/storage"

	"github.com/gin-gonic/gin"
)

// main 主函数。
func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	// 初始化分词组件。
	if err := common.Initialize(); err != nil {
		slog.Error("[INI]加载分词组件失败", "error", err)
		return
	}
	// 初始化存储层。
	if err := storage.Initialize(); err != nil {
		slog.Error("[INI]初始化存储层失败", "error", err)
		return
	}
	defer storage.Dispose()
	// 初始化服务层。
	service.Initialize()
	// 初始化处理层。
	handler.Initialize()
	// 启动任务调度。
	pipeline.Initialize()
	defer pipeline.Dispose()
	// 设置运行模式。
	switch config.Cfg.Station.Mode {
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
	address := fmt.Sprintf("0.0.0.0:%d", config.Cfg.Station.Port)
	slog.Info("[INI]正在启动监听服务", "address", address)
	if err := engine.Run(address); err != nil {
		slog.Error("[INI]启动监听服务失败", "error", err)
	}
}
