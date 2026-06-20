package infra

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-inventory/internal/config"
)

func NewRedisClient(cfg config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("[infra] Redis connected")
	return rdb, nil
}
