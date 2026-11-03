package service

// Worker 后台任务执行器。
//
// 负责取任务、控制并发、执行处理流程并回写进度。
type Worker struct{}

// NewWorker 创建后台任务执行器。
func NewWorker() *Worker {
	return &Worker{}
}
