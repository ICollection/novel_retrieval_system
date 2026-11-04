package server

import (
	"net/http"
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

// TestLogin 验证用户登录。
func TestLogin(ctx *gin.Context) {
	buffer := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims, err := ParseToken(buffer)
	if err != nil {
		Failure(ctx, Failure401, err.Error())
		return
	}
	ctx.Set("claims", claims)
	ctx.Next()
}

// TestPermit 验证用户权限。
func TestPermit(permission string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet("claims").(*Claims)
		if claims.HasPermission(permission) == false {
			Failure(ctx, Failure403, "用户缺少权限。")
			return
		}
		ctx.Next()
	}
}
