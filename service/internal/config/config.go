package config

// Config 应用配置。
type Config struct {
	Embedding  Embedding  `mapstructure:"embedding"`
	PostgreSQL PostgreSQL `mapstructure:"postgresql"`
	Qdrant     Qdrant     `mapstructure:"qdrant"`
	MinIO      MinIO      `mapstructure:"minio"`
}

// Embedding 向量化服务配置。
type Embedding struct {
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

// PostgreSQL 数据库连接配置。
type PostgreSQL struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
}

// Qdrant 向量数据库连接配置。
type Qdrant struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	APIKey     string `mapstructure:"api_key"`
	Collection string `mapstructure:"collection"`
}

// MinIO 对象存储配置。
type MinIO struct {
	Endpoint  string `mapstructure:"endpoint"`
	Port      int    `mapstructure:"port"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Region    string `mapstructure:"region"`
}
