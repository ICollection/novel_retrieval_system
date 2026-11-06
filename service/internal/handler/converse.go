package handler

import (
	"log"
	"service/internal/common"
	"service/internal/config"
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
	// 保存用户消息。
	session_id, message_id, _, err := service.Converse.AppendHumanMessage(ctx, session, content, novels)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	log.Printf("会话编码：%s，消息编码：%s", session_id, message_id)
	// 生成系统消息。
	answer := "aaaaa"
	// 保存系统消息。
	// 返回成功结果。
	server.Success(ctx, answer)
}
