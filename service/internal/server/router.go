package server

import (
	"net/http"

	"service/internal/handler"
	"service/internal/server/response"
)

// NewRouter 创建路由并挂载中间件。
func NewRouter(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{"status": "ok"})
	})

	// 业务一：上传文件与处理进度。
	mux.HandleFunc("POST /api/v1/files", h.Upload)
	mux.HandleFunc("GET /api/v1/files", h.List)
	mux.HandleFunc("GET /api/v1/files/{id}/progress", h.Progress)
	mux.HandleFunc("POST /api/v1/files/{id}/retry", h.Retry)
	mux.HandleFunc("DELETE /api/v1/files/{id}", h.Delete)

	// 业务二：对话与查询。
	mux.HandleFunc("POST /api/v1/chat", h.Chat)

	return Chain(mux, Recover, Logger, CORS)
}
