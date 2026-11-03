package qdrant

import "service/internal/config"

// Store Qdrant 存储，封装向量写入与检索。
type Store struct {
	host       string
	port       int
	apiKey     string
	collection string
}

// New 创建 Qdrant 存储。
func New() *Store {
	cfg := config.Config.Qdrant
	return &Store{host: cfg.Host, port: cfg.Port, apiKey: cfg.APIKey, collection: cfg.Collection}
}
