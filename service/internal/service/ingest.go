package service

// Ingest 文件接入与处理业务。
//
// 负责上传文件、创建处理任务、维护与查询处理进度。
type Ingest struct{}

// NewIngest 创建文件接入业务。
func NewIngest() *Ingest {
	return &Ingest{}
}
