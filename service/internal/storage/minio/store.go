package minio

import "service/internal/config"

// Store MinIO 存储，封装文件对象的上传、下载与删除。
type Store struct {
	cfg config.MinIO
}

// New 创建 MinIO 存储。
func New(cfg config.MinIO) *Store {
	return &Store{cfg: cfg}
}
