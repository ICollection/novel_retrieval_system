package model

import "time"

// 处理状态，对应 progress.status。
const (
	StatusPending   = "01" // 待处理
	StatusParsing   = "02" // 解析中
	StatusChunking  = "03" // 切片中
	StatusEmbedding = "04" // 向量化中
	StatusDone      = "05" // 已完成
	StatusFailed    = "06" // 已失败
)

// Origin 原始文件。
type Origin struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Size       int64     `json:"size"`
	Hash       string    `json:"hash"`
	CreateTime time.Time `json:"create_time"`
}

// Novel 小说信息。
type Novel struct {
	ID             string    `json:"id"`
	Origin         string    `json:"origin"`
	Title          string    `json:"title"`
	CharacterCount int       `json:"character_count"`
	ChapterCount   int       `json:"chapter_count"`
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
}

// Chapter 章节信息。
type Chapter struct {
	ID             string `json:"id"`
	Novel          string `json:"novel"`
	Serial         int    `json:"serial"`
	Title          string `json:"title"`
	Start          int    `json:"start"`
	Final          int    `json:"final"`
	CharacterCount int    `json:"character_count"`
}

// Progress 处理进度。
type Progress struct {
	ID         string `json:"id"`
	Novel      string `json:"novel"`
	Status     string `json:"status"`
	TotalCount int    `json:"total_count"`
	EmbedCount int    `json:"embed_count"`
}

// FileItem 前端文件列表项。
type FileItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Status     string `json:"status"`
	TotalCount int    `json:"total_count"`
	EmbedCount int    `json:"embed_count"`
	CreateTime string `json:"create_time"`
}
