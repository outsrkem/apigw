#!/bin/bash

[ -d output ] || mkdir output
# 比对项目文件中引入的依赖与go.mod进行比对,清理不需要的依赖,并且更新go.sum文件。
go mod tidy
# 将项目的所有依赖导出至vendor目录
go mod vendor
# 构建
go build -o output/apigw src/main/main.go
