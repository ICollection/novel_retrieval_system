package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Result 返回结果。
type Result struct {
	Success bool    `json:"success"`
	Failure int     `json:"failure"`
	Message *string `json:"message,omitempty"`
	Content any     `json:"content,omitempty"`
}

// Success 返回成功结果。
func Success(ctx *gin.Context, content any) {
	ctx.AbortWithStatusJSON(http.StatusOK, Result{Success: true, Failure: Success200, Message: nil, Content: content})
}

// Failure 返回失败结果。
func Failure(ctx *gin.Context, failure int, message string) {
	ctx.AbortWithStatusJSON(http.StatusOK, Result{Success: false, Failure: failure, Message: &message, Content: nil})
}

// Success200 请求结果成功。
var Success200 int = 200

// Failure400 请求参数错误。
var Failure400 int = 400

// Failure401 用户没有登录。
var Failure401 int = 401

// Failure403 用户没有权限。
var Failure403 int = 403

// Failure501 一般业务错误。
var Failure501 int = 501
