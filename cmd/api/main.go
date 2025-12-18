package main

import (
	"fmt"
	"seckill-system/internal/database"
	"seckill-system/internal/handler"
	"seckill-system/internal/middleware"
	mqPkg "seckill-system/internal/pkg/mq"
	redisPkg "seckill-system/internal/pkg/redis"
	"seckill-system/internal/service"
	"seckill-system/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// 加载配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./internal/config")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	//初始化数据库
	db := database.InitMySQL()
	//初始化redis
	redisPkg.InitRedis()
	//初始化RabbitMQ
	mqPkg.InitRabbitMQ()
	defer mqPkg.Close()

	//启动订单消费者
	orderConsumer := &service.OrderConsumer{
		DB: db,
	}
	orderConsumer.Start()

	// 初始化 Gin
	r := gin.Default()

	//创建Product处理器实例
	productService := &service.ProductService{
		DB: db,
	}
	//启动回灌：DB-->Redis库存协程
	if err := productService.SyncStockToRedis(); err != nil {
		panic(fmt.Errorf("sync stock to redis failed: %s", err))
	}
	productHandler := &handler.ProductHandler{
		ProductService: productService,
	}

	//创建User处理器实例
	userHandler := &handler.UserHandler{
		DB: db,
	}

	//用户注册登录
	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)

	//需要认证的路由
	auth := r.Group("/user")
	auth.Use(middleware.Auth())
	{
		auth.GET("/info", func(c *gin.Context) {
			uid := c.GetUint("uid")
			c.JSON(200, gin.H{"uid": uid})
		})
	}

	//秒杀相关路由
	SeckillService := &service.SeckillService{}

	r.POST("/product", productHandler.Create)
	r.GET("/products", productHandler.List)

	auth.POST("/seckill/:id", func(c *gin.Context) {
		uid := c.GetUint("uid")
		id := utils.StrToUint(c.Param("id"))
		if id == 0 {
			c.JSON(400, gin.H{"error": "invalid product id"})
			return
		}

		err := SeckillService.StartSeckill(id, uid)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "秒杀成功"})
	})

	// 🔧 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		health := gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		}

		// 检查Redis
		if _, err := redisPkg.RDB.Ping(redisPkg.Ctx).Result(); err != nil {
			health["redis"] = "unhealthy: " + err.Error()
			c.JSON(503, health)
			return
		}
		health["redis"] = "ok"

		// 检查MySQL
		sqlDB, _ := db.DB()
		if err := sqlDB.Ping(); err != nil {
			health["mysql"] = "unhealthy: " + err.Error()
			c.JSON(503, health)
			return
		}
		health["mysql"] = "ok"

		// 检查RabbitMQ
		if mqPkg.Conn == nil || mqPkg.Conn.IsClosed() {
			health["rabbitmq"] = "unhealthy: connection closed"
			c.JSON(503, health)
			return
		}
		health["rabbitmq"] = "ok"

		c.JSON(200, health)
	})

	// 🔧 监控接口：获取系统统计信息
	r.GET("/stats", func(c *gin.Context) {
		stats := gin.H{
			"timestamp": time.Now().Unix(),
		}

		// Redis统计
		redisInfo, _ := redisPkg.RDB.Info(redisPkg.Ctx, "stats").Result()
		stats["redis"] = redisInfo

		// MySQL统计
		sqlDB, _ := db.DB()
		dbStats := sqlDB.Stats()
		stats["mysql"] = gin.H{
			"open_connections": dbStats.OpenConnections,
			"in_use":           dbStats.InUse,
			"idle":             dbStats.Idle,
		}

		c.JSON(200, stats)
	})

	// 启动服务器
	fmt.Println("🚀 Server starting on :8080")
	fmt.Println("   Health: http://localhost:8080/health")
	fmt.Println("   Stats:  http://localhost:8080/stats")
	r.Run(":8080")

	// 启动服务器
	r.Run(":8080")
}
