package handler

import (
	"net/http"

	"service/internal/server/response"
)

// Chat 对话提问。
// TODO 绑定 model.ChatRequest，调用 service 完成检索与生成。
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	response.NotImplemented(w)
}
