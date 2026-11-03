package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// Body 统一响应体。
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// OK 返回成功响应。
func OK(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, Body{Code: 0, Message: "ok", Data: data})
}

// Error 返回错误响应。
func Error(w http.ResponseWriter, status int, message string) {
	write(w, status, Body{Code: status, Message: message})
}

// NotImplemented 返回接口未实现响应。
func NotImplemented(w http.ResponseWriter) {
	Error(w, http.StatusNotImplemented, "接口尚未实现")
}

func write(w http.ResponseWriter, status int, body Body) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Println("写入响应失败，详细错误信息：", err)
	}
}
