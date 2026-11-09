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
//
// param: engine 网站引擎。
func (ac *accountHandler) register(engine *gin.Engine) {
	group := engine.Group(config.Cfg.Station.ContextPath + "/account")
	group.GET("/filterAccount", server.TestLogin, ac.filterAccount)
	group.POST("/appendAccount", server.TestPermit(common.PermissionManageUser), ac.appendAccount)
	group.PATCH("/modifyAccount", server.TestPermit(common.PermissionManageUser), ac.modifyAccount)
	group.PATCH("/changeAccount", server.TestPermit(common.PermissionManageUser), ac.changeAccount)
	group.DELETE("/deleteAccount", server.TestPermit(common.PermissionManageUser), ac.deleteAccount)
	group.GET("/verifyAccount", ac.verifyAccount)
	group.GET("/consumeAccount", server.TestPermit(common.PermissionManageUser), ac.consumeAccount)
}

// filterAccount 查询系统用户。
//
// param: ctx 上下文。
func (ac *accountHandler) filterAccount(ctx *gin.Context) {
	// 调用服务方法。
	result, err := service.Account.FilterAccount(ctx)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}

// appendAccount 添加系统用户。
//
// param: ctx 上下文。
func (ac *accountHandler) appendAccount(ctx *gin.Context) {
	// 获取输入参数。
	name := ac.getString(ctx, "name")
	login := ac.getString(ctx, "login")
	password := ac.getString(ctx, "password")
	permission := ac.getString(ctx, "permission")
	// 验证输入参数。
	if name == nil || login == nil || password == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.AppendAccount(ctx, name, login, password, permission)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// modifyAccount 修改系统用户。
//
// param: ctx 上下文。
func (ac *accountHandler) modifyAccount(ctx *gin.Context) {
	// 获取输入参数。
	id := ac.getString(ctx, "id")
	name := ac.getString(ctx, "name")
	permission := ac.getString(ctx, "permission")
	// 验证输入参数。
	if id == nil || name == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.ModifyAccount(ctx, id, name, permission)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// changeAccount 改密系统用户。
//
// param: ctx 上下文。
func (ac *accountHandler) changeAccount(ctx *gin.Context) {
	// 获取输入参数。
	id := ac.getString(ctx, "id")
	password := ac.getString(ctx, "password")
	// 验证输入参数。
	if id == nil || password == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.ChangeAccount(ctx, id, password)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// deleteAccount 删除系统用户。
//
// param: ctx 上下文。
func (ac *accountHandler) deleteAccount(ctx *gin.Context) {
	// 获取输入参数。
	id := ac.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.DeleteAccount(ctx, id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// verifyAccount 验证系统用户。
//
// param: ctx 上下文。
func (ac *accountHandler) verifyAccount(ctx *gin.Context) {
	// 获取输入参数。
	login := ac.getString(ctx, "login")
	password := ac.getString(ctx, "password")
	// 验证输入参数。
	if login == nil || password == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.VerifyAccount(ctx, login, password)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// consumeAccount 统计用户词元。
//
// param: ctx 上下文。
func (ac *accountHandler) consumeAccount(ctx *gin.Context) {
	// 获取输入参数。
	id := ac.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := service.Account.ConsumeAccount(ctx, id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}
