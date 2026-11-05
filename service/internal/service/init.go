package service

// Account 系统用户服务。
var AccountService *accountService = nil

// Origin 原始文件服务。
var OriginService *originService = nil

// Initalize 初始化服务层。
func Initalize() {
	AccountService = &accountService{}
	OriginService = &originService{}
}
