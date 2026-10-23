package db

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis() *redis.Client {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "localhost:6379"
	}

	return redis.NewClient(&redis.Options{
		Addr: url,
	})
}

func CloseRedis(client *redis.Client) {
	if client != nil {
		_ = client.Close()
	}
}
