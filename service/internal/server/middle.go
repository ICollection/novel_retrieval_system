package server

import (
	"net/http"
	"service/internal/config"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CrossDomain 允许跨域访问。
func CrossDomain(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	ctx.Header("Access-Control-Expose-Headers", "X-Token")
	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
	ctx.Next()
}

// TestLogin 验证用户登录。
//
// param: ctx 上下文。
func TestLogin(ctx *gin.Context) {
	// 验证用户令牌。
	buffer := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims, err := ParseToken(buffer)
	if err != nil {
		Failure(ctx, Failure401, err.Error())
		return
	}
	ctx.Set("claims", claims)
	// 更新用户令牌。
	if claims.ExpiresAt.Sub(time.Now()).Hours() < float64(config.Cfg.Station.Expire/2) {
		token, err := IssueToken(&claims.Subject, claims.Name, claims.Login, claims.Permission)
		if err == nil {
			ctx.Header("X-Token", *token)
		}
	}
	// 继续处理下文。
	ctx.Next()
}

// TestPermit 验证用户权限。
//
// param: permission 权限需求。
// return: 处理函数。
func TestPermit(permission string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 验证用户令牌。
		buffer := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
		if buffer == "" {
			Failure(ctx, Failure401, "用户没有登录。")
			return
		}
		claims, err := ParseToken(buffer)
		if err != nil {
			Failure(ctx, Failure401, err.Error())
			return
		}
		// 验证用户权限。
		if claims.HasPermission(permission) == false {
			Failure(ctx, Failure403, "用户缺少权限。")
			return
		}
		ctx.Set("claims", claims)
		// 更新用户令牌。
		if claims.ExpiresAt.Sub(time.Now()).Hours() < float64(config.Cfg.Station.Expire/2) {
			token, err := IssueToken(&claims.Subject, claims.Name, claims.Login, claims.Permission)
			if err == nil {
				ctx.Header("X-Token", *token)
			}
		}
		// 继续处理下文。
		ctx.Next()
	}
}
