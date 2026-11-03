package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// content 编译期嵌入的配置文件内容。
//
//go:embed config.yaml
var content []byte

// Config 全局配置。
var Config *config

// init 初始化全局配置。
func init() {
	cfg, err := parse()
	if err != nil {
		log.Fatal("初始化配置失败，详细错误信息：", err)
	}
	Config = cfg
}

// parse 解析内嵌的配置文件。
func parse() (*config, error) {
	reader := viper.New()
	reader.SetConfigType("yaml")
	if err := reader.ReadConfig(bytes.NewReader(content)); err != nil {
		return nil, fmt.Errorf("解析配置文件失败，详细错误信息：%w", err)
	}

	cfg := &config{}
	if err := reader.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("反序列化配置文件失败，详细错误信息：%w", err)
	}
	return cfg, nil
}
