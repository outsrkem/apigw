package redis

import (
	"apigw/src/cfgtypes"
	"apigw/src/slog"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	maxRetries = 100
	maxBackoff = 10 * time.Second
	PoolSize   = 100
)

var Rdb *redis.Client

func InitRedis(cfg *cfgtypes.Redis) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	klog := slog.FromCtx(ctx)
	klog.Info("Connect redis: ", cfg.Addr)

	retries := 0
	backoff := time.Second

	for retries < maxRetries {
		cli := redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.Db,
			PoolSize: PoolSize,
		})

		// Check if the connection is successful.
		pong, err := cli.Ping(context.Background()).Result()
		if err != nil {
			_ = cli.Close()

			retries++
			klog.Errorf("Failed to connect to redis, retry count:%d, retrying in %v... err:%s", retries, backoff, err)
			time.Sleep(backoff)

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		Rdb = cli
		klog.Infof("redis connection is successful. %s", pong)
		return
	}

	klog.Fatalf("redis connect failed after %d retries", maxRetries)
}
