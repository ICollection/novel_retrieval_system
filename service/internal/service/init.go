package service

var (
	// Account 系统用户服务。
	Account *accountService = nil
	// Origin 原始文件服务。
	Origin *originService = nil
	// Novel 小说对象服务。
	Novel *novelService = nil
	// Mission 分析任务服务。
	Mission *missionService = nil
	// Converse 用户对话服务。
	Converse *converseService = nil
)

// Initialize 初始化服务层。
func Initialize() {
	Account = &accountService{}
	Origin = &originService{}
	Novel = &novelService{}
	Mission = &missionService{}
	Converse = &converseService{}
}
