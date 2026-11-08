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
	group := engine.Group(config.Cfg.Station.ContextPath + "/converse")
	group.GET("/filterSession", server.TestLogin, handler.filterSession)
	group.POST("/modifySession", server.TestLogin, handler.modifySession)
	group.DELETE("/deleteSession", server.TestLogin, handler.deleteSession)
	group.GET("/filterMessage", server.TestLogin, handler.filterMessage)
	group.POST("/appendMessage", server.TestLogin, handler.appendMessage)
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

// modifySession 修改用户会话。
//
// param: ctx 上下文。
func (handler *converseHandler) modifySession(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	title := handler.getString(ctx, "title")
	// 验证输入参数。
	if id == nil || title == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Converse.ModifySession(ctx, *id, *title)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// deleteSession 删除用户会话。
//
// param: ctx 上下文。
func (handler *converseHandler) deleteSession(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Converse.DeleteSession(ctx, *id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// filterMessage 查询会话消息。
func (handler *converseHandler) filterMessage(ctx *gin.Context) {
	// 获取输入参数。
	session := handler.getString(ctx, "session")
	// 验证输入参数。
	if session == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Converse.FilterRecentMessage(ctx, session, 20)
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
		return
	}
	// 返回成功结果。
	writer.Success(*sessionId)
}
