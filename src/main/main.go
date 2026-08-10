// src/main/main.go
package main

import (
	"apigw/src/cache"
	"apigw/src/config"
	"apigw/src/core/helpers"
	"apigw/src/core/service"
	"apigw/src/database/mysql"
	"apigw/src/database/redis"
	"apigw/src/pkg/session"
	"apigw/src/route"
	"apigw/src/slog"
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	cfg := config.InitConfig()

	slog.InitLogger(&cfg.Apigw.Log)
	var klog = slog.GetGlobal()
	klog.Info("start apigw service...")

	app := cfg.Apigw.App
	rcfg := cfg.Apigw.Redis
	auth := cfg.Apigw.Auth

	mysql.InitDB(&cfg.Apigw)
	redis.InitRedis(&cfg.Apigw.Redis)

	// Initialize gateway runtime cache instance and set as global singleton
	cacheIns := cache.NewGatewayCache()
	cache.SetGlobalCache(cacheIns)

	// Initialize global route matching service instance
	service.RouteSvc = service.NewRouteService(cacheIns)

	// Load full route data from database into cache on service startup
	err := helpers.LoadAndRefreshRoute(cacheIns)
	if err != nil {
		klog.Fatal("init route load failed", err)
	}

	// Load full channel configuration from database into cache on service startup
	err = helpers.LoadAndRefreshChannel(cacheIns)
	if err != nil {
		klog.Fatal("init channel load failed", err)
	}

	// Start background goroutine for redis pub/sub configuration hot-reload watcher
	ctxWatch := context.Background()
	go helpers.StartRouteWatcher(ctxWatch, cacheIns)

	// Create hertz http server instance with basic parameters
	svc := server.Default(
		server.WithHostPorts(app.Bind),
		server.WithMaxRequestBodySize(200<<20), // Max request body limit: 200MB
		server.WithRedirectTrailingSlash(false),
		server.WithExitWaitTime(0*time.Second))

	route.Middleware(svc)
	session.InitSession(svc, &rcfg)
	route.ApigwRoute(svc)
	route.AuthRouter(svc, &auth)
	route.ProxyRoute(svc)

	// Block and start http service
	svc.Spin()
}
