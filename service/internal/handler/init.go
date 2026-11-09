package handler

import (
	"github.com/gin-gonic/gin"
)

var (
	// account 系统用户处理器。
	account *accountHandler = nil
	// prepare 准备资料处理器。
	prepare *prepareHandler = nil
	// discuss 用户问答处理器。
	discuss *discussHandler = nil
)

// Initialize 初始化处理器层。
func Initialize() {
	account = &accountHandler{}
	prepare = &prepareHandler{}
	discuss = &discussHandler{}
}

// Register 注册处理器路由。
func Register(engine *gin.Engine) {
	account.register(engine)
	prepare.register(engine)
	discuss.register(engine)
}
