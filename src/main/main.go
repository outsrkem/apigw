/*
 * Copyright
 */

package main

import (
	"apigw/src/config"
	"apigw/src/pkg/session"
	"apigw/src/route"
	"apigw/src/slog"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	cfg := config.InitConfig() // 解析配置文件
	slog.InitLogger(&cfg.Apigw.Log)
	klog := slog.FromContext(nil)
	klog.Info("start apigw service.")
	app := cfg.Apigw.App
	proxy := cfg.Apigw.Proxy
	redis := cfg.Apigw.Redis
	auth := cfg.Apigw.Auth

	svc := server.Default(server.WithHostPorts(app.Bind), server.WithExitWaitTime(0*time.Second))
	route.Middleware(svc)
	session.InitSession(svc, &redis)
	route.LocalRouter(svc, &auth)
	route.ProxyRouter(svc, &proxy)
	svc.Spin()
}
