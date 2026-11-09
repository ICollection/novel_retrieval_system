package handler

import (
	"service/internal/common"
	"service/internal/config"
	"service/internal/server"
	"service/internal/service"

	"github.com/gin-gonic/gin"
)

// discussHandler 用户问答处理器。
type discussHandler struct {
	handler
}

// register 注册路由分组。
func (dh *discussHandler) register(engine *gin.Engine) {
	group := engine.Group(config.Cfg.Station.ContextPath + "/discuss")
	group.GET("/filterSession", server.TestLogin, dh.filterSession)
	group.POST("/modifySession", server.TestLogin, dh.modifySession)
	group.DELETE("/deleteSession", server.TestLogin, dh.deleteSession)
	group.GET("/filterMessage", server.TestLogin, dh.filterMessage)
	group.POST("/appendMessage", server.TestLogin, dh.appendMessage)
}

// filterSession 查询用户对话。
func (dh *discussHandler) filterSession(ctx *gin.Context) {
	// 调用服务方法。
	result, err := service.Discuss.FilterSession(ctx)
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
func (dh *discussHandler) modifySession(ctx *gin.Context) {
	// 获取输入参数。
	id := dh.getString(ctx, "id")
	title := dh.getString(ctx, "title")
	// 验证输入参数。
	if id == nil || title == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Discuss.ModifySession(ctx, *id, *title)
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
func (dh *discussHandler) deleteSession(ctx *gin.Context) {
	// 获取输入参数。
	id := dh.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Discuss.DeleteSession(ctx, *id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// filterMessage 查询会话消息。
//
// param: ctx 上下文。
func (dh *discussHandler) filterMessage(ctx *gin.Context) {
	// 获取输入参数。
	session := dh.getString(ctx, "session")
	recent := dh.getInt32(ctx, "recent")
	// 验证输入参数。
	if session == nil || recent == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Discuss.FilterMessage(ctx, session, *recent)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// appendMessage 添加对话消息
//
// param: ctx 上下文。
func (dh *discussHandler) appendMessage(ctx *gin.Context) {
	// 获取输入参数。
	session := dh.getString(ctx, "session")
	content := dh.getString(ctx, "content")
	novels := common.Split(dh.getString(ctx, "novels"), ",")
	// 验证输入参数。
	if content == nil || len(novels) == 0 {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	writer := server.CreateAnswerWriter(ctx)
	// 保存用户消息。
	sessionId, err := service.Compose.ComposeAnswer(ctx, session, content, novels, writer.Process)
	if err != nil {
		writer.Failure(err.Error())
		return
	}
	// 返回成功结果。
	writer.Success(*sessionId)
}
