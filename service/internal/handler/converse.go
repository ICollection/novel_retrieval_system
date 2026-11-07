package handler

import (
	"service/internal/common"
	"service/internal/config"
	"service/internal/response"
	"service/internal/server"
	"service/internal/service"

	"github.com/gin-gonic/gin"
)

// 用户对话服务。
type converseHandler struct {
	handler
}

// register 注册路由分组。
func (handler *converseHandler) register(engine *gin.Engine) {
	group := engine.Group(config.Config.Station.ContextPath + "/converse")
	group.POST("/filterSession", server.TestLogin, handler.filterSession)
}

// filterSession 查询用户对话。
func (handler *converseHandler) filterSession(ctx *gin.Context) {
	// 调用服务方法。
	result, err := service.Converse.FilterSession(ctx)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// appendMessage 添加对话消息
func (handler *converseHandler) appendMessage(ctx *gin.Context) {
	// 获取输入参数。
	session := handler.getString(ctx, "session")
	content := handler.getString(ctx, "content")
	novels := common.Split(handler.getString(ctx, "novels"), ",")
	// 验证输入参数。
	if content == nil || len(novels) == 0 {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	writer := server.CreateAnswerWriter(ctx)
	// 保存用户消息。
	sessionId, err := response.ComposeAnswer(ctx, session, content, novels, writer.Process)
	if err != nil {
		writer.Failure(err.Error())
	}
	// 返回成功结果。
	writer.Success(*sessionId)
}
