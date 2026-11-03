package config

import "log"

// Conf 全局配置。
var Conf *Config

// init 初始化全局配置。
func init() {
	config, err := Parse()
	if err != nil {
		log.Fatal("初始化配置失败，详细错误信息：", err)
	}
	Conf = config
}
