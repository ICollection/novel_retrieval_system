package service

var (
	// Account 系统用户服务。
	Account *accountService = nil
	// Prepare 准备资料服务。
	Prepare *prepareService = nil
	// Converse 用户对话服务。
	Converse *converseService = nil
)

// Initialize 初始化服务层。
func Initialize() {
	Account = &accountService{}
	Prepare = &prepareService{}
	Converse = &converseService{}
}
