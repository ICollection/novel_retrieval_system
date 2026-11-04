package service

// Account 系统用户服务。
var Account *AccountService = nil

// init 初始化服务层。
func init() {
	account, _ := createAccountService()
	Account = account
}
