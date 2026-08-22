# Changelog
All notable changes to this project will be documented in this file.

格式遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]
### 待更新

- 暂无变更条目

## [0.2.0] - 2026-09-7
### 新增
* 使用MySQL存储接口数据
* 网关后台管理接口
* 支持按分组绑定独立域名
* 支持通过独立域名接入
* 支持会话空闲超时，优化令牌续期

## 变更
* 重构网关能力
* 支持数据库接口配置，优化性能

## 依赖
* uias >= 0.3.1

## [0.1.0] - 2026-08-05
### 新增
* 日志输出到文件
* 接口匹配模式

### 修复
* 升级 hertz 依赖至 0.9.7
* 修复接口处理逻辑

## [0.0.1] - 2024-09-24
> 初始版本发布
### 新增
* 接口代理
* 接口提权
