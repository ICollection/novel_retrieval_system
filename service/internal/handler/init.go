package handler

import (
	"github.com/gin-gonic/gin"
)

// account 系统用户处理器。
var account *AccountHandler = nil

// init 初始化处理器层。
func init() {
	account, _ = createAccountHandler()
}

// Register 注册处理器路由。
func Register(engine *gin.Engine) {
	account.register(engine)
}
