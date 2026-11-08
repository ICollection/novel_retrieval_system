package storage

import "service/internal/config"

var (
	// RDB 关系数据库连接。
	RDB *PostgreClient
	// VDB 向量数据库连接。
	VDB *QdrantClient
	// OSS 对象存储连接。
	OSS *MinioClient
	// LLM 大语言模型连接。
	LLM *ModelClient
	// SLM 小语言模型连接。
	SLM *ModelClient
	// VEM 向量模型连接。
	VEM *ModelClient
)

// Initialize 初始化存储层。
func Initialize() error {
	var err error
	// 初始化关系数据库连接。
	if RDB, err = newPostgreClient(config.Cfg.Postgres); err != nil {
		return err
	}
	// 初始化向量数据库连接。
	if VDB, err = newQdrantClient(config.Cfg.Qdrant); err != nil {
		return err
	}
	if err = VDB.CreateCollection(); err != nil {
		return err
	}
	// 初始化对象存储连接。
	if OSS, err = newMinioClient(config.Cfg.MinIO); err != nil {
		return err
	}
	// 初始化大语言模型连接。
	if LLM, err = newModelClient(config.Cfg.Models.Large, "[LLM]"); err != nil {
		return err
	}
	// 初始化小语言模型连接。
	if SLM, err = newModelClient(config.Cfg.Models.Small, "[SLM]"); err != nil {
		return err
	}
	// 初始化向量模型连接。
	if VEM, err = newModelClient(config.Cfg.Models.Embed, "[VEM]"); err != nil {
		return err
	}
	// 返回成功结果。
	return nil
}

// Dispose 回收存储层。
func Dispose() {
	if RDB != nil {
		RDB.dispose()
		RDB = nil
	}
}
