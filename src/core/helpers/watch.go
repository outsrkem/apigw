package helpers

import (
	"apigw/src/cache"
	"apigw/src/database/redis"
	"apigw/src/slog"
	"context"
	"time"
)

// routeRefreshChannel redis pub/sub channel for route dynamic refresh notification
const routeRefreshChannel = "gw:route:refresh"

// WatchRouteChange listens redis pub/sub channel, reload route & channel cache when receiving signal
// Trigger command in Redis CLI: PUBLISH gw:route:refresh reload
func WatchRouteChange(ctx context.Context, c *cache.GatewayCache) {
	r := redis.Rdb
	sub := r.Subscribe(ctx, routeRefreshChannel)
	klog := slog.FromCtx(ctx)
	// Close subscription gracefully and record close error
	defer func() {
		closeErr := sub.Close()
		if closeErr != nil {
			klog.Errorf("redis route refresh subscribe close error: %v", closeErr)
		}
	}()

	ch := sub.Channel()
	klog.Info("route & channel refresh watcher running")

	for {
		select {
		case <-ctx.Done():
			klog.Info("route refresh watcher exit, context canceled")
			return
		case _, ok := <-ch:
			if !ok {
				klog.Error("redis subscribe channel closed abnormally, watcher exit")
				return
			}
			klog.Info("receive refresh signal, reloading routes and channels")

			err := LoadAndRefreshRoute(c)
			if err != nil {
				klog.Errorf("reload route failed: %v", err)
			}

			err = LoadAndRefreshChannel(c)
			if err != nil {
				klog.Errorf("reload channel failed: %v", err)
			} else {
				klog.Info("all cache refresh success")
			}
		}
	}
}

// StartRouteWatcher wrapper function for automatic reconnection after subscription disconnection
func StartRouteWatcher(ctx context.Context, c *cache.GatewayCache) {
	klog := slog.FromCtx(ctx)
	for {
		WatchRouteChange(ctx, c)
		klog.Warn("route refresh watcher exited, restart after 1 second")
		time.Sleep(time.Second)
	}
}
