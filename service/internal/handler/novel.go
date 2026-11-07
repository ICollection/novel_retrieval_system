package handler

import (
	"service/internal/common"
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
	group.GET("/filter", server.TestLogin, handler.filter)
	group.PATCH("/modify", server.TestPermit(common.PermissionManageOrigin), handler.modify)
	group.DELETE("/delete", server.TestPermit(common.PermissionManageOrigin), handler.delete)
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

// modify 修改小说对象。
func (handler *novelHandler) modify(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	title := handler.getString(ctx, "title")
	// 验证输入参数。
	if id == nil || title == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Novel.Modify(ctx, id, title)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}
func (handler *novelHandler) delete(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Novel.Delete(ctx, id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}
