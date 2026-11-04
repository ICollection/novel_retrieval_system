package service

// Account 系统用户服务。
var Account *AccountService = nil

// Origin 原始文件服务。
var Origin *OriginService = nil

// init 初始化服务层。
func init() {
	Account, _ = createAccountService()
	Origin, _ = createOriginService()
}
