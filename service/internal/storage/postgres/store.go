package postgres

import "service/internal/config"

// Store PostgreSQL 存储，封装连接与原子读写。
type Store struct {
	cfg config.PostgreSQL
}

// New 创建 PostgreSQL 存储。
func New(cfg config.PostgreSQL) *Store {
	return &Store{cfg: cfg}
}
