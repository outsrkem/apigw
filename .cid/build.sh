#!/bin/bash

[ -d output ] || mkdir output
# 比对项目文件中引入的依赖与go.mod进行比对,清理不需要的依赖,并且更新go.sum文件。
go mod tidy
# 将项目的所有依赖导出至vendor目录
go mod vendor
# 构建
LD_PATH="apigw/src/config"
APIGW_VERSION="0.0.1"
GO_VERSION=$(go version |awk '{print $3}')
flags="-X $LD_PATH.Version=${APIGW_VERSION} -X $LD_PATH.GoVersion=${GO_VERSION} -X $LD_PATH.GitCommit=123"
go build -trimpath -ldflags "-s -w $flags" -o output/apigw src/main/main.go
