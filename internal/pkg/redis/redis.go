package redis

import (
	"context"
	"time"

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

		// 🔧 连接池配置
		PoolSize:     100, // 连接池大小，根据并发量调整
		MinIdleConns: 10,  // 最小空闲连接，保持热连接

		// 🔧 超时配置
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读超时
		WriteTimeout: 3 * time.Second, // 写超时

		// 🔧 重试配置
		MaxRetries:      3, // 最大重试次数
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,

		// 🔧 连接生命周期
		PoolTimeout: 4 * time.Second, // 获取连接的超时时间
		IdleTimeout: 5 * time.Minute, // 空闲连接超时

		// 🔧 健康检查（可选）
		// OnConnect: func(ctx context.Context, cn *redis.Conn) error {
		//     return cn.Ping(ctx).Err()
		// },
	})

	// 测试连接
	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}

	println("✅ Redis initialized successfully")
	println("   - Pool Size: 100")
	println("   - Min Idle Conns: 10")
}
