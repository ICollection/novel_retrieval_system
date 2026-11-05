package handler

import (
	"service/internal/common"
	"service/internal/config"
	"service/internal/server"
	"service/internal/service"

	"github.com/gin-gonic/gin"
)

// originHandler 原始文件处理器。
type originHandler struct {
	handler
}

// register 注册路由分组。
func (handler *originHandler) register(engine *gin.Engine) {
	group := engine.Group(config.Config.Station.ContextPath + "/origin")
	group.POST("/upload", server.TestPermit(common.PermissionManageOrigin), handler.upload)
}

// upload 上传原始文件。
func (handler *originHandler) upload(ctx *gin.Context) {
	// 获取输入参数。
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Origin.Upload(ctx, file, header)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}
