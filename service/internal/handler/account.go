package handler

import (
	"service/internal/common"
	"service/internal/config"
	"service/internal/server"
	"service/internal/service"

	"github.com/gin-gonic/gin"
)

// accountHandler 系统用户处理器。
type accountHandler struct {
	handler
}

// register 注册路由分组。
func (handler *accountHandler) register(engine *gin.Engine) {
	group := engine.Group(config.Config.Station.ContextPath + "/account")
	group.POST("/append", server.TestPermit(common.PermissionManageUser), handler.append)
	group.PATCH("/modify", server.TestPermit(common.PermissionManageUser), handler.modify)
	group.PATCH("/change", server.TestPermit(common.PermissionManageUser), handler.change)
	group.DELETE("/delete", server.TestPermit(common.PermissionManageUser), handler.delete)
	group.GET("/verify", handler.verify)
	group.GET("/filter", server.TestLogin, handler.filter)
}

// append 添加系统用户。
func (handler *accountHandler) append(ctx *gin.Context) {
	// 获取输入参数。
	name := handler.getString(ctx, "name")
	login := handler.getString(ctx, "login")
	password := handler.getString(ctx, "password")
	permission := handler.getString(ctx, "permission")
	// 验证输入参数。
	if name == nil || login == nil || password == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.Append(ctx, name, login, password, permission)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// modify 修改系统用户。
func (handler *accountHandler) modify(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	name := handler.getString(ctx, "name")
	permission := handler.getString(ctx, "permission")
	// 验证输入参数。
	if id == nil || name == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.Modify(ctx, id, name, permission)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// change 改密系统用户。
func (handler *accountHandler) change(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	password := handler.getString(ctx, "password")
	// 验证输入参数。
	if id == nil || password == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.Change(ctx, id, password)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// delete 删除系统用户。
func (handler *accountHandler) delete(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.Delete(ctx, id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// verify 验证系统用户。
func (handler *accountHandler) verify(ctx *gin.Context) {
	// 获取输入参数。
	login := handler.getString(ctx, "login")
	password := handler.getString(ctx, "password")
	// 验证输入参数。
	if login == nil || password == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.Verify(ctx, login, password)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// filter 查询系统用户。
func (handler *accountHandler) filter(ctx *gin.Context) {
	// 调用服务方法。
	result, err := service.Account.Filter(ctx)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}
