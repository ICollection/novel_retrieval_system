package config

import (
	"net"
	"net/url"
	"strconv"
)

// Config 应用配置。
type Config struct {
	Station  Station  `mapstructure:"station"`
	Models   Models   `mapstructure:"models"`
	Postgres Postgres `mapstructure:"postgres"`
	Qdrant   Qdrant   `mapstructure:"qdrant"`
	MinIO    Minio    `mapstructure:"minio"`
}

// Station 站点与登录令牌配置。
type Station struct {
	Port        int    `mapstructure:"port"`
	ContextPath string `mapstructure:"context_path"`
	Mode        string `mapstructure:"mode"`
	Secret      string `mapstructure:"secret"`
	Expire      int    `mapstructure:"expire"`
	Concurrency int    `mapstructure:"concurrency"`
}

// Models 模型服务配置。
type Models struct {
	Embed Model `mapstructure:"embed"`
	Large Model `mapstructure:"large"`
	Small Model `mapstructure:"small"`
}

// Model 各类模型配置。
type Model struct {
	Address string `mapstructure:"address"`
	Name    string `mapstructure:"name"`
	Key     string `mapstructure:"key"`
	Timeout int    `mapstructure:"timeout"`
}

// Postgres 数据库连接配置。
type Postgres struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
	Timeout  int    `mapstructure:"timeout"`
}

// GetDSN 获取连接字符串。
//
// return: 连接字符串。
func (p *Postgres) GetDSN() string {
	url := url.URL{Scheme: "postgres", User: url.UserPassword(p.Username, p.Password), Host: net.JoinHostPort(p.Host, strconv.Itoa(p.Port)), Path: "/" + p.Database, RawQuery: "sslmode=" + p.SSLMode}
	return url.String()
}

// Qdrant 向量数据库连接配置。
type Qdrant struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	APIKey     string `mapstructure:"api_key"`
	Collection string `mapstructure:"collection"`
	Timeout    int    `mapstructure:"timeout"`
}

// Minio 对象存储配置。
type Minio struct {
	Endpoint  string `mapstructure:"endpoint"`
	Port      int    `mapstructure:"port"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Region    string `mapstructure:"region"`
	Timeout   int    `mapstructure:"timeout"`
}
