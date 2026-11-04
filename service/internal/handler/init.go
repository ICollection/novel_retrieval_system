package handler

import (
	"github.com/gin-gonic/gin"
)

// account 系统用户处理器。
var account *AccountHandler = nil

// origin 原始文件处理器。
var origin *OriginHandler = nil

// init 初始化处理器层。
func init() {
	account, _ = createAccountHandler()
	origin, _ = createOriginHandler()
}

// Register 注册处理器路由。
func Register(engine *gin.Engine) {
	account.register(engine)
	origin.register(engine)
}
