package handler

import (
	"github.com/gin-gonic/gin"
)

var (
	// account 系统用户处理器。
	account *accountHandler = nil
	// origin 原始文件处理器。
	origin *originHandler = nil
	// novel 小说对象处理器。
	novel *novelHandler = nil
	// mission 分析任务处理器。
	mission *missionHandler = nil
)

// Initalize 初始化处理器层。
func Initalize() {
	account = &accountHandler{}
	origin = &originHandler{}
	novel = &novelHandler{}
	mission = &missionHandler{}
}

// Register 注册处理器路由。
func Register(engine *gin.Engine) {
	account.register(engine)
	origin.register(engine)
	novel.register(engine)
	mission.register(engine)
}
