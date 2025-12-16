package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var (
	RDB *redis.Client          // Redis 客户端实例
	Ctx = context.Background() // Redis 上下文
)

func InitRedis() {
	addr := viper.GetString("redis.addr")

	RDB = redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
}
