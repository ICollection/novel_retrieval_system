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
//
// param: ctx 上下文。
// param: key 参数名称。
// return: 整数参数。
func (h *handler) getString(ctx *gin.Context, key string) *string {
	value := strings.TrimSpace(ctx.Request.FormValue(key))
	if value == "" {
		return nil
	}
	return &value
}

// getInt 获取整数参数。
//
// param: ctx 上下文。
// param: key 参数名称。
// return: 整数参数。
func (h *handler) getInt(ctx *gin.Context, key string) *int {
	value, err := strconv.Atoi(ctx.Request.FormValue(key))
	if err != nil {
		return nil
	}
	return &value
}

// getInt32 获取整数参数。
//
// param: ctx 上下文。
// param: key 参数名称。
// return: 整数参数。
func (h *handler) getInt32(ctx *gin.Context, key string) *int32 {
	value, err := strconv.ParseInt(ctx.Request.FormValue(key), 10, 32)
	if err != nil {
		return nil
	}
	result := int32(value)
	return &result
}

// getInt64 获取整数参数。
//
// param: ctx 上下文。
// param: key 参数名称。
// return: 整数参数。
func (h *handler) getInt64(ctx *gin.Context, key string) *int64 {
	value, err := strconv.ParseInt(ctx.Request.FormValue(key), 10, 64)
	if err != nil {
		return nil
	}
	return &value
}

// register 注册路由接口。
type register interface {
	Register(engine *gin.Engine)
}
