package config

import (
	"net"
	"net/url"
	"strconv"
)

// config 应用配置。
type config struct {
	Station  station  `mapstructure:"station"`
	Models   models   `mapstructure:"models"`
	Postgres postgres `mapstructure:"postgres"`
	Qdrant   qdrant   `mapstructure:"qdrant"`
	MinIO    minio    `mapstructure:"minio"`
}

// station 站点与登录令牌配置。
type station struct {
	Port        int    `mapstructure:"port"`
	ContextPath string `mapstructure:"context_path"`
	Mode        string `mapstructure:"mode"`
	Secret      string `mapstructure:"secret"`
	Expire      int    `mapstructure:"expire"`
	Concurrency int    `mapstructure:"concurrency"`
}

// models 模型服务配置。
type models struct {
	Embed model `mapstructure:"embed"`
	Large model `mapstructure:"large"`
	Small model `mapstructure:"small"`
}

// model 各类模型配置。
type model struct {
	Address string `mapstructure:"address"`
	Name    string `mapstructure:"name"`
	Key     string `mapstructure:"key"`
	Timeout int    `mapstructure:"timeout"`
}

// postgres 数据库连接配置。
type postgres struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
	Timeout  int    `mapstructure:"timeout"`
}

// GetDSN 获取连接字符串。
func (postgres *postgres) GetDSN() string {
	url := url.URL{Scheme: "postgres", User: url.UserPassword(postgres.Username, postgres.Password), Host: net.JoinHostPort(postgres.Host, strconv.Itoa(postgres.Port)), Path: "/" + postgres.Database, RawQuery: "sslmode=" + postgres.SSLMode}
	return url.String()
}

// qdrant 向量数据库连接配置。
type qdrant struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	APIKey     string `mapstructure:"api_key"`
	Collection string `mapstructure:"collection"`
	Timeout    int    `mapstructure:"timeout"`
}

// minio 对象存储配置。
type minio struct {
	Endpoint  string `mapstructure:"endpoint"`
	Port      int    `mapstructure:"port"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Region    string `mapstructure:"region"`
	Timeout   int    `mapstructure:"timeout"`
}
