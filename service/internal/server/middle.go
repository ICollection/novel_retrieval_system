package server

import (
	"net/http"
	"service/internal/server/response"
	"service/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

// CrossDomain 允许跨域访问。
func CrossDomain(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
	ctx.Next()
}

// Identity 获取登录用户。
func Identity(ctx *gin.Context) {
	buffer := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims, err := service.ParseToken(buffer)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		ctx.Abort()
		return
	}
	ctx.Set("claims", claims)
	ctx.Next()
}
