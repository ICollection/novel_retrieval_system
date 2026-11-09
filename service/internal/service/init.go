package service

var (
	// Account 系统用户服务。
	Account *accountService
	// Prepare 准备资料服务。
	Prepare *prepareService
	// Discuss 用户问答服务。
	Discuss *discussService
	// Compose 问答编排服务。
	Compose *composeService
)

// Initialize 初始化服务层。
func Initialize() {
	Account = &accountService{}
	Prepare = &prepareService{}
	Discuss = &discussService{}
	Compose = &composeService{}
}
