package storage

import (
	"fmt"
	"service/internal/config"

	qdrants "github.com/qdrant/go-client/qdrant"
)

// qdrant Qdrant 向量数据库连接。
type qdrant struct {
	client *qdrants.Client
}

// createQdrant 创建 Qdrant 向量数据库连接。
func createQdrant() (*qdrant, error) {
	client, err := qdrants.NewClient(&qdrants.Config{Host: config.Config.Qdrant.Host, Port: config.Config.Qdrant.Port, APIKey: config.Config.Qdrant.APIKey})
	if err != nil {
		return nil, fmt.Errorf("创建 Qdrant 向量数据库连接失败，详细错误信息：%w", err)
	}
	return &qdrant{client: client}, nil
}
