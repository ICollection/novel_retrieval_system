package handler

import (
	"service/internal/config"
	"service/internal/server"
	"service/internal/service"

	"github.com/gin-gonic/gin"
)

// novelHandler 小说对象处理器。
type novelHandler struct {
	handler
}

// register 注册路由分组。
func (handler *novelHandler) register(engine *gin.Engine) {
	group := engine.Group(config.Config.Station.ContextPath + "/novel")
	group.POST("/filter", server.TestLogin, handler.filter)
}

// filter 查询小说对象。
func (handler *novelHandler) filter(ctx *gin.Context) {
	// 调用服务方法。
	result, err := service.Novel.Filter(ctx)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}
