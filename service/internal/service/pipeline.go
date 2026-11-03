package service

// Pipeline 文本处理流程：解析章节 → 切片 → 向量化。
//
// 只包含纯处理逻辑，不直接读写数据库，便于单独测试。
type Pipeline struct{}

// NewPipeline 创建处理流程。
func NewPipeline() *Pipeline {
	return &Pipeline{}
}
