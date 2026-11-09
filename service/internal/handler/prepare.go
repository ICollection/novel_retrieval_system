package handler

import (
	"service/internal/common"
	"service/internal/config"
	"service/internal/server"
	"service/internal/service"

	"github.com/gin-gonic/gin"
)

// prepareHandler 准备资料处理器。
type prepareHandler struct {
	handler
}

// register 注册路由分组。
//
// param: engine 网站引擎。
func (ph *prepareHandler) register(engine *gin.Engine) {
	group := engine.Group(config.Cfg.Station.ContextPath + "/prepare")
	group.POST("/uploadOrigin", server.TestPermit(common.PermissionManageOrigin), ph.uploadOrigin)
	group.GET("/filterMission", server.TestLogin, ph.filterMission)
	group.PATCH("/launchMission", server.TestPermit(common.PermissionManageMission), ph.launchMission)
	group.PATCH("/cancelMission", server.TestPermit(common.PermissionManageMission), ph.cancelMission)
	group.GET("/filterNovel", server.TestLogin, ph.filterNovel)
	group.PATCH("/modifyNovel", server.TestPermit(common.PermissionManageOrigin), ph.modifyNovel)
	group.DELETE("/deleteNovel", server.TestPermit(common.PermissionManageOrigin), ph.deleteNovel)
}

// uploadOrigin 上传原始文件。
//
// param: ctx 上下文。
func (ph *prepareHandler) uploadOrigin(ctx *gin.Context) {
	// 获取输入参数。
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Prepare.UploadOrigin(ctx, file, header)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// filterMission 查询分析任务。
//
// param: ctx 上下文。
func (ph *prepareHandler) filterMission(ctx *gin.Context) {
	// 调用服务方法。
	result, err := service.Prepare.FilterMission(ctx)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// launchMission 启动分析任务。
//
// param: ctx 上下文。
func (ph *prepareHandler) launchMission(ctx *gin.Context) {
	// 获取输入参数。
	id := ph.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Prepare.LaunchMission(ctx, id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// cancelMission 取消分析任务。
//
// param: ctx 上下文。
func (ph *prepareHandler) cancelMission(ctx *gin.Context) {
	// 获取输入参数。
	id := ph.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Prepare.CancelMission(ctx, id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// filterNovel 查询小说对象。
//
// param: ctx 上下文。
func (ph *prepareHandler) filterNovel(ctx *gin.Context) {
	// 调用服务方法。
	result, err := service.Prepare.FilterNovel(ctx)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// modifyNovel 修改小说对象。
//
// param: ctx 上下文。
func (ph *prepareHandler) modifyNovel(ctx *gin.Context) {
	// 获取输入参数。
	id := ph.getString(ctx, "id")
	title := ph.getString(ctx, "title")
	// 验证输入参数。
	if id == nil || title == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Prepare.ModifyNovel(ctx, id, title)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// deleteNovel 删除小说对象。
//
// param: ctx 上下文。
func (ph *prepareHandler) deleteNovel(ctx *gin.Context) {
	// 获取输入参数。
	id := ph.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Prepare.DeleteNovel(ctx, id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}
