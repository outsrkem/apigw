package redis

// https://blog.51cto.com/u_16213356/9902194
// https://www.cnblogs.com/guyouyin123/p/14481650.html

import (
	"apigw/src/cfgtypes"
	"apigw/src/slog"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client

// InitRedis initializes a connection to a Redis server.
func InitRedis(cfg *cfgtypes.Redis) {
	var klog = slog.GetGlobal()
	klog.Info("Connect redis: ", cfg.Addr)
	klog.Debug("Redis passwd: ", cfg.Password)
	retries := 0
	backoff := time.Second
	var ctx context.Context
	var cancel context.CancelFunc
	for {
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

		// Create a new Redis client.
		Rdb = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,     // Redis server address.
			Password: cfg.Password, // Redis password.
			DB:       cfg.Db,       // Redis database number.
			PoolSize: 100,          // Maximum number of connections in the pool.
		})

		// Check if the connection is successful.
		pong, err := Rdb.Ping(ctx).Result()
		if err == nil {
			klog.Info("redis conPossible resource leak, 'defer' is called in the 'for' loopnection is successful. ", pong)
			break
		}

		retries++
		if retries >= 100 {
			panic(err)
		}

		klog.Errorf("Failed to connect to redis. Retrying in %v...", backoff)
		time.Sleep(backoff)

		backoff += time.Second
	}
	defer cancel()
}
