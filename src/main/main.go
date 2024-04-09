package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	"apigw/src/config"
	"apigw/src/mw"
	"apigw/src/route"
)

func main() {

	// 解析配置文件
	cfgApigw := config.InitConfig()
	proxy := cfgApigw.Apigw.Rroxy
	redis := cfgApigw.Apigw.Redis

	// 初始化hertz
	h := server.Default(server.WithHostPorts("0.0.0.0:8080"))
	mw.InitSession(h, &redis)
	route.RouteLocal(h)
	route.RouteProxy(h, &proxy)
	// 启动服务
	h.Spin()
}
