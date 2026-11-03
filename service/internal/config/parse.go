package config

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/spf13/viper"
)

// content 编译期嵌入的配置文件内容。
//
//go:embed config.yaml
var content []byte

// Parse 解析内嵌的配置文件。
func Parse() (*Config, error) {
	reader := viper.New()
	reader.SetConfigType("yaml")
	if err := reader.ReadConfig(bytes.NewReader(content)); err != nil {
		return nil, fmt.Errorf("解析配置文件失败，详细错误信息：%w", err)
	}

	config := &Config{}
	if err := reader.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("反序列化配置文件失败，详细错误信息：%w", err)
	}
	return config, nil
}
