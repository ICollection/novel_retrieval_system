package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

// content 编译期嵌入的配置文件内容。
//
//go:embed config.yaml
var content []byte

// Recall 召回用系统提示词。
//
//go:embed recall.txt
var Recall string

// Normal 问答用系统提示词。
//
//go:embed normal.txt
var Normal string

// Compress 压缩用系统提示词。
//
//go:embed compress.txt
var Compress string

// Cfg 全局配置。
var Cfg *Config

// init 初始化全局配置。
func init() {
	cfg, err := parse()
	if err != nil {
		log.Fatal("解析系统配置失败，详细错误信息：", err)
	}
	Cfg = cfg
}

// parse 解析配置文件。
//
// return: 全局配置，错误信息。
func parse() (*Config, error) {
	// 读取配置文件。
	reader := viper.New()
	reader.SetConfigType("yaml")
	if err := reader.ReadConfig(bytes.NewReader(content)); err != nil {
		return nil, fmt.Errorf("读取配置文件失败，详细错误信息：%w", err)
	}
	// 解析配置文件。
	cfg := &Config{}
	if err := reader.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败，详细错误信息：%w", err)
	}
	// 补充环境变量。
	if cfg.Models.Embed.Key == "" {
		cfg.Models.Embed.Key = os.Getenv("NOVEL_EMBED_KEY")
	}
	if cfg.Models.Large.Key == "" {
		cfg.Models.Large.Key = os.Getenv("NOVEL_LARGE_KEY")
	}
	if cfg.Models.Small.Key == "" {
		cfg.Models.Small.Key = os.Getenv("NOVEL_SMALL_KEY")
	}
	// 返回配置信息。
	return cfg, nil
}
