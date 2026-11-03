package handler

import (
	"github.com/gin-gonic/gin"

	"service/internal/server/response"
	"service/internal/service"
)

// ChatHandler 对话与查询接口。
type ChatHandler struct {
	chat *service.ChatService
}

// NewChatHandler 创建对话接口。
func NewChatHandler(chat *service.ChatService) *ChatHandler {
	return &ChatHandler{chat: chat}
}

// Register 注册对话相关路由。
func (h *ChatHandler) Register(group *gin.RouterGroup) {
	group.POST("/chat", h.Chat)
}

// Chat 对话提问。
// TODO 绑定 model.ChatRequest，调用 service 完成检索与生成。
func (h *ChatHandler) Chat(c *gin.Context) {
	response.NotImplemented(c)
}
