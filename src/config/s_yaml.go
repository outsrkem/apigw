package config

// Config yaml配置结构体
type Config struct {
	Apigw Apigw `yaml:"apigw"`
}

type Apigw struct {
	App   App     `yaml:"app"`
	Redis Redis   `yaml:"redis"`
	Rroxy []Proxy `yaml:"proxy"`
}

type App struct {
	Bind string `yaml:"bind"`
}

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	Db       string `yaml:"db"`
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
