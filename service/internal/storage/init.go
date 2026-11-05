package storage

var (
	// RDB 关系数据库连接。
	RDB *postgres
	// VDB 向量数据库连接。
	VDB *qdrant
	// OSS 对象存储连接。
	OSS *minio
	// LLM 语言模型连接。
	LLM *openai
	// VMD 向量模型连接。
	VMD *openai
)

// Initalize 初始化存储层。
func Initalize() error {
	var err error
	// 初始化关系数据库连接。
	if RDB, err = createPostgres(); err != nil {
		return err
	}
	// 初始化向量数据库连接。
	if VDB, err = createQdrant(); err != nil {
		return err
	}
	// 初始化对象存储连接。
	if OSS, err = createMinio(); err != nil {
		return err
	}
	// 初始化语言模型连接。
	if LLM, err = createOpenaiLLM(); err != nil {
		return err
	}
	// 初始化向量模型连接。
	if VMD, err = createOpenaiVMD(); err != nil {
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
