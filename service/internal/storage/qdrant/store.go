package qdrant

import "service/internal/config"

// Store Qdrant 存储，封装向量写入与检索。
type Store struct {
	cfg config.Qdrant
}

// New 创建 Qdrant 存储。
func New(cfg config.Qdrant) *Store {
	return &Store{cfg: cfg}
}
