package model

// ChatRequest 提问请求。
type ChatRequest struct {
	Question string `json:"question"`
}

// ChatReply 回答响应。
type ChatReply struct {
	Answer  string   `json:"answer"`
	Sources []Source `json:"sources"`
}

// Source 回答引用的原文片段。
type Source struct {
	Chapter string `json:"chapter"`
	Title   string `json:"title"`
	Text    string `json:"text"`
}
