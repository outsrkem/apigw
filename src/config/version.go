package config

import (
	"fmt"
	"os"
)

var (
	Version   = "" // Version 项目版本信息
	GoVersion = "" // GoVersion Go版本信息
	GitCommit = "" // GitCommit git提交commmit id
)

type versions struct {
	AppVersion string
	GoVersion  string
	GitCommit  string
}

func newVersions(appv, gov, commit string) (*versions, error) {
	v := &versions{
		AppVersion: appv,
		GoVersion:  gov,
		GitCommit:  commit,
	}
	return v, nil
}

func (v *versions) Print(versions *versions) {
	fmt.Println("Version: ", versions.AppVersion)
	fmt.Println("Go Version: ", versions.GoVersion)
	fmt.Println("Git Commit: ", versions.GitCommit)
	os.Exit(0)
}
