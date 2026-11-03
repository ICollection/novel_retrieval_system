package handler

import "service/internal/service"

// Handler 接口层，负责参数解析、校验与响应输出。
type Handler struct {
	ingest *service.Ingest
	chat   *service.Chat
}

// New 创建接口层。
func New(ingest *service.Ingest, chat *service.Chat) *Handler {
	return &Handler{ingest: ingest, chat: chat}
}
