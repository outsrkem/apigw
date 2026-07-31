package config

import (
	"apigw/src/cfgtypes"
	"flag"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"os"

	"gopkg.in/yaml.v2"
)

type FlagArgs struct {
	CfgPath      string
	PrintVersion bool
	Plain        string
}

func NewFlagArgs() *FlagArgs {
	fa := &FlagArgs{}
	flag.StringVar(&fa.CfgPath, "c", "apigw.yaml", "Configuration file path")
	flag.BoolVar(&fa.PrintVersion, "version", false, "print program version")
	flag.StringVar(&fa.Plain, "encrypt", "", "Encrypted string.")
	flag.Parse()
	return fa
}

func InitConfig() *cfgtypes.Config {
	var _cfg cfgtypes.Config
	fa := NewFlagArgs()

	if fa.PrintVersion {
		versions, _ := newVersions(Version, GoVersion, GitCommit)
		versions.Print(versions)
	}
	if fa.Plain != "" {
		encryption(fa.Plain)
	}

	hlog.Infof("Read configuration file: %s", fa.CfgPath)
	configData, err := os.ReadFile(fa.CfgPath)
	if err != nil {
		hlog.Errorf("Read configuration file error: %v", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(configData, &_cfg)
	if err != nil {
		hlog.Errorf("Unmarshal configuration file error: %v", err)
		os.Exit(1)
	}

	decryCfgCipher(&_cfg.Apigw.Database.Passwd)
	decryCfgCipher(&_cfg.Apigw.Redis.Password)
	return &_cfg
}
