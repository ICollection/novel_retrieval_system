package handler

import (
	"errors"
	"service/internal/config"
	"service/internal/server"
	"service/internal/service"

	"github.com/gin-gonic/gin"
)

// OriginHandler 原始文件处理器。
type OriginHandler struct {
	originService *service.OriginService
	handler
}

// createOriginHandler 创建原始文件处理器。
func createOriginHandler() (*OriginHandler, error) {
	if service.Account == nil {
		return nil, errors.New("无法创建系统用户服务。")
	}
	return &OriginHandler{originService: service.Origin}, nil
}

// register 注册路由分组。
func (handler *OriginHandler) register(engine *gin.Engine) {
	group := engine.Group(config.Config.Web.ContextPath + "/origin")
	group.POST("/upload", server.TestPermit(server.UploadNovel), handler.upload)
}

// upload 上传原始文件。
func (handler *OriginHandler) upload(ctx *gin.Context) {
	// 获取输入参数。
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := handler.originService.Upload(ctx, file, header)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}
