package config

import (
	"net"
	"net/url"
	"strconv"
)

// config 应用配置。
type config struct {
	Web       web       `mapstructure:"web"`
	Embedding embedding `mapstructure:"embedding"`
	Postgre   postgre   `mapstructure:"postgre"`
	Qdrant    qdrant    `mapstructure:"qdrant"`
	MinIO     minio     `mapstructure:"minio"`
}

// web 服务配置。
type web struct {
	Port        int    `mapstructure:"port"`
	ContextPath string `mapstructure:"context_path"`
	Mode        string `mapstructure:"mode"`
}

// embedding 向量化服务配置。
type embedding struct {
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

// postgre 数据库连接配置。
type postgre struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
}

// GetDSN 获取连接字符串。
func (postgre *postgre) GetDSN() string {
	url := url.URL{Scheme: "postgres", User: url.UserPassword(postgre.Username, postgre.Password), Host: net.JoinHostPort(postgre.Host, strconv.Itoa(postgre.Port)), Path: "/" + postgre.Database, RawQuery: "sslmode=" + postgre.SSLMode}
	return url.String()
}

// qdrant 向量数据库连接配置。
type qdrant struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	APIKey     string `mapstructure:"api_key"`
	Collection string `mapstructure:"collection"`
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
}
