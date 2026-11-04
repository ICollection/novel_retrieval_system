package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// handler 通用处理器。
type handler struct {
}

// getString 获取字符传参数。
func (handler *handler) getString(ctx *gin.Context, key string) *string {
	value := strings.TrimSpace(ctx.Request.FormValue(key))
	if value == "" {
		return nil
	}
	return &value
}

// getInt 获取整数参数。
func (handler *handler) getInt(ctx *gin.Context, key string) *int {
	value, err := strconv.Atoi(ctx.Request.FormValue(key))
	if err != nil {
		return nil
	}
	return &value
}

// register 注册路由接口。
type register interface {
	Register(engine *gin.Engine)
}
