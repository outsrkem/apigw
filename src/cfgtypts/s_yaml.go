package cfgtypts

type Config struct {
	Apigw Apigw `yaml:"apigw"`
}

type Apigw struct {
	App   App     `yaml:"app"`
	Redis Redis   `yaml:"redis"`
	Auth  Auth    `yaml:"auth"`
	Log   Log     `yaml:"log"`
	Proxy []Proxy `yaml:"proxy"`
}

type App struct {
	Bind string `yaml:"bind"`
}

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	Db       string `yaml:"db"`
}

type Auth struct {
	Backend Backend `yaml:"backend"`
}

type Log struct {
	Level  string `yaml:"level"`
	Output Output `yaml:"output"`
}

type Output struct {
	File   File   `yaml:"file"`
	Stdout string `yaml:"stdout"`
}

type File struct {
	Name       string `yaml:"name"`
	MaxSize    int    `yaml:"maxsize"`
	MaxBackups int    `yaml:"maxbackups"`
	MaxAge     int    `yaml:"maxage"`
	Compress   bool   `yaml:"compress"`
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
