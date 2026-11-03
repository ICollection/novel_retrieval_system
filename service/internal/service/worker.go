package service

// WorkerService 后台任务执行器。
//
// 负责取任务、控制并发、执行处理流程并回写进度。
type WorkerService struct{}

// NewWorkerService 创建后台任务执行器。
func NewWorkerService() *WorkerService {
	return &WorkerService{}
}
