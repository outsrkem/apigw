package config

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Config 配置结构体
type Config struct {
	Apigw Apigw `yaml:"apigw"`
}

type Apigw struct {
	Redis Redis   `yaml:"redis"`
	Rroxy []Proxy `yaml:"proxy"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
}

type Proxy struct {
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
	flag.StringVar(&cfgPath, "f", "config.yaml", "Configuration file path")
	flag.Parse()
	log.Println("Read configuration file:", cfgPath)

	configData, err := os.ReadFile(cfgPath)
	if err != nil {
		log.Println("读取配置文件失败:", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(configData, &_cfg)
	if err != nil {
		log.Println("解析配置文件失败:", err)
		os.Exit(1)
	}
	// proxy := _cfg.Apigw.Rroxy
	// redis := _cfg.Apigw.Redis
	// for _, apigw := range proxy {
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
