package redisx

import "github.com/go-redis/redis/v8"

var Manager *RedisManager

// RedisManager unifies all redis repository instances
type RedisManager struct {
	client     *redis.Client
	AgwSession *AgwSessionRepo
}

// NewRedisManager creates RedisManager, pass initialized *redis.Client
func NewRedisManager(client *redis.Client) *RedisManager {
	mgr := &RedisManager{
		client: client,
	}
	mgr.AgwSession = NewAgwSessionRepo(client)
	return mgr
}
