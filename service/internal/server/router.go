package server

import (
	"strings"

	"github.com/gin-gonic/gin"

	"service/internal/server/response"
)

// Registrar 可注册路由的业务接口。
type Registrar interface {
	Register(group *gin.RouterGroup)
}

// Setup 设置 Gin 运行模式、挂载中间件并注册路由。
// contextPath 为业务接口的路径前缀，mode 为 Gin 运行模式。
func Setup(engine *gin.Engine, contextPath string, mode string, handlers ...Registrar) {
	setMode(mode)
	engine.Use(gin.Logger(), gin.Recovery(), CORS())

	engine.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	group := engine.Group(normalize(contextPath))
	for _, h := range handlers {
		h.Register(group)
	}
}

// normalize 规范化上下文目录：确保以 / 开头，且不以 / 结尾。
func normalize(contextPath string) string {
	prefix := strings.TrimSpace(contextPath)
	prefix = strings.TrimSuffix(prefix, "/")
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return prefix
}

// setMode 设置 Gin 运行模式，非法值按 debug 处理。
func setMode(mode string) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case gin.ReleaseMode:
		gin.SetMode(gin.ReleaseMode)
	case gin.TestMode:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}
