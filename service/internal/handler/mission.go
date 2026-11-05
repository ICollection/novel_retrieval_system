package handler

import (
	"service/internal/common"
	"service/internal/config"
	"service/internal/server"
	"service/internal/service"

	"github.com/gin-gonic/gin"
)

// missionHandler 分析任务处理器。
type missionHandler struct {
	handler
}

// register 注册路由分组。
func (handler *missionHandler) register(engine *gin.Engine) {
	group := engine.Group(config.Config.Station.ContextPath + "/mission")
	group.POST("/filter", server.TestLogin, handler.filter)
	group.POST("/launch", server.TestPermit(common.PermissionManageMission), handler.launch)
	group.POST("/cancel", server.TestPermit(common.PermissionManageMission), handler.cancel)
}

// filter 查询分析任务。
func (handler *missionHandler) filter(ctx *gin.Context) {
	// 调用服务方法。
	result, err := service.Mission.Filter(ctx)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// launch 启动分析任务。
func (handler *missionHandler) launch(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Mission.Launch(ctx, id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// cancel 取消分析任务。
func (handler *missionHandler) cancel(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Mission.Cancel(ctx, id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}
