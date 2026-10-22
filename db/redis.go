package db

import "github.com/redis/go-redis/v9"

var Redis *redis.Client

func ConnectRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func CloseRedis() {
	if Redis != nil {
		_ = Redis.Close()
	}
}
