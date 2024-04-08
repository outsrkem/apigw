package config

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config 配置结构体
type Config struct {
	Apigw []Apigw `yaml:"apigw"`
}

type Apigw struct {
	Name   string   `yaml:"name"`
	Server []Server `yaml:"server"`
}

type Server struct {
	Location Location `yaml:"location"`
}

type Location struct {
	Path    string  `yaml:"path"`
	Backend Backend `yaml:"backend"`
}

type Backend struct {
	Host string `yaml:"host"`
	Url  string `yaml:"url"`
}

// InitConfig 初始化配置
func InitConfig() *Config {
	var _cfg Config
	var cfgPath string
	flag.StringVar(&cfgPath, "f", "config/config.yaml", "Configuration file path")
	flag.Parse()
	fmt.Println("Read configuration file:", cfgPath)

	configData, err := os.ReadFile(cfgPath)
	if err != nil {
		fmt.Println("读取配置文件失败:", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(configData, &_cfg)
	if err != nil {
		fmt.Println("解析配置文件失败:", err)
		os.Exit(1)
	}
	// for _, apigw := range _cfg.Apigw {
	//     fmt.Println("")
	//     fmt.Println(apigw.Name)
	//     for _, server := range apigw.Server {
	//         fmt.Println(server.Location.Path)
	//         fmt.Println(server.Location.Backend.Host)
	//         fmt.Println(server.Location.Backend.Url)
	//     }

	// }

	return &_cfg
}
