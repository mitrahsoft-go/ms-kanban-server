package redis

import (
	"context"
	"fmt"

	"github.com/ms-kanban-server/config/configs"
	"github.com/redis/go-redis/v9"
)

func InitRedisClient(cfg *configs.Config) (*redis.Client, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})
	defer rdb.Close()

	// The client Ping method sends the real Redis PING command
	err := PingRedis(rdb)
	if err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}
	return rdb, err
}

func PingRedis(rdb *redis.Client) error {
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	return nil
}
