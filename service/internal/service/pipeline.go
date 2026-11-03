package service

// PipelineService 文本处理流程：解析章节 → 切片 → 向量化。
//
// 只包含纯处理逻辑，不直接读写数据库，便于单独测试。
type PipelineService struct{}

// NewPipelineService 创建处理流程。
func NewPipelineService() *PipelineService {
	return &PipelineService{}
}
