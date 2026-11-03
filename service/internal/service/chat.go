package service

// Chat 对话与查询业务。
//
// 负责检索相关片段、拼装上下文并调用模型生成回答。
type Chat struct{}

// NewChat 创建对话业务。
func NewChat() *Chat {
	return &Chat{}
}
