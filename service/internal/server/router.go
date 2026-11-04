package server

import (
	"github.com/gin-gonic/gin"
)

// Register 可注册路由的业务接口。
type Register interface {
	Register(group *gin.RouterGroup)
}
