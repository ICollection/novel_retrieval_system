package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body 统一响应体。
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: "ok", Data: data})
}

// Error 返回错误响应并终止后续处理。
func Error(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, Body{Code: status, Message: message})
}

// NotImplemented 返回接口未实现响应。
func NotImplemented(c *gin.Context) {
	Error(c, http.StatusNotImplemented, "接口尚未实现")
}
