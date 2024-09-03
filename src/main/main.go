/*
 * Copyright
 */

package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	"apigw/src/config"
	"apigw/src/mw"
	"apigw/src/pkg/session"
	"apigw/src/route"
)

func init() {

}

func main() {

	// 解析配置文件
	cfgApigw := config.InitConfig()
	app := cfgApigw.Apigw.App
	proxy := cfgApigw.Apigw.Rroxy
	redis := cfgApigw.Apigw.Redis
	auth := cfgApigw.Apigw.Auth

	// 初始化hertz
	// app.Bind 监听参数：default 127.0.0.1:8080
	h := server.Default(server.WithHostPorts(app.Bind))
	mw.NewAccessLog(h)
	session.InitSession(h, &redis)
	route.RouteLocal(h, &auth)
	route.RouteProxy(h, &proxy)
	// 启动服务
	h.Spin()
}
