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

	// Parse and load global configuration file
	cfg := config.InitConfig()
	// Initialize global logging component
	slog.InitLogger(&cfg.Apigw.Log)
	var klog = slog.GetGlobal()
	klog.Info("start apigw service.")

	app := cfg.Apigw.App
	rcfg := cfg.Apigw.Redis
	auth := cfg.Apigw.Auth

	// Initialize MySQL database connection pool
	mysql.InitDB(&cfg.Apigw)
	// Initialize Redis client connection
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
		server.WithExitWaitTime(0*time.Second))

	// Register global middleware chain
	route.Middleware(svc)
	// Initialize distributed session component
	session.InitSession(svc, &rcfg)
	// Register gateway management routes
	route.ApigwRoute(svc)
	// Register local internal service routes
	route.AuthRouter(svc, &auth)
	// Register core proxy forwarding business routes
	route.ProxyRoute(svc, &auth)

	// Block and start http service
	svc.Spin()
}
