package handler

import (
	"errors"
	"service/internal/config"
	"service/internal/server"
	"service/internal/service"

	"github.com/gin-gonic/gin"
)

// AccountHandler 系统用户处理器。
type AccountHandler struct {
	accountService *service.AccountService
	handler
}

// createAccountHandler 创建系统用户处理器。
func createAccountHandler() (*AccountHandler, error) {
	if service.Account == nil {
		return nil, errors.New("无法创建系统用户服务。")
	}
	return &AccountHandler{accountService: service.Account}, nil
}

// Register 注册路由分组。
func (handler *AccountHandler) Register(engine *gin.Engine) {
	group := engine.Group(config.Config.Web.ContextPath + "/account")
	group.POST("/append", handler.append)
	group.PATCH("/modify", handler.modify)
	group.PATCH("/change", handler.change)
	group.DELETE("/delete", handler.delete)
	group.GET("/verify", handler.verify)
	group.GET("/filter", handler.filter)
}

// append 添加系统用户。
func (handler *AccountHandler) append(ctx *gin.Context) {
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
	result, err := handler.accountService.Append(ctx, name, login, password, permission)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// modify 修改系统用户。
func (handler *AccountHandler) modify(ctx *gin.Context) {
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
	result, err := handler.accountService.Modify(ctx, id, name, permission)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// change 改密系统用户。
func (handler *AccountHandler) change(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	password := handler.getString(ctx, "password")
	// 验证输入参数。
	if id == nil || password == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := handler.accountService.Change(ctx, id, password)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// delete 删除系统用户。
func (handler *AccountHandler) delete(ctx *gin.Context) {
	// 获取输入参数。
	id := handler.getString(ctx, "id")
	// 验证输入参数。
	if id == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := handler.accountService.Delete(ctx, id)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// verify 验证系统用户。
func (handler *AccountHandler) verify(ctx *gin.Context) {
	// 获取输入参数。
	login := handler.getString(ctx, "login")
	password := handler.getString(ctx, "password")
	// 验证输入参数。
	if login == nil || password == nil {
		server.Failure(ctx, server.Failure400, "请求参数错误。")
		return
	}
	// 调用服务方法。
	result, err := handler.accountService.Verify(ctx, login, password)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, *result)
}

// filter 查询系统用户。
func (handler *AccountHandler) filter(ctx *gin.Context) {
	// 调用服务方法。
	result, err := handler.accountService.Filter(ctx)
	if err != nil {
		server.Failure(ctx, server.Failure501, err.Error())
		return
	}
	// 返回成功结果。
	server.Success(ctx, result)
}
