package database

import (
	"fmt"
	"log"
	"seckill-system/internal/model"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitMySQL() *gorm.DB {
	user := viper.GetString("mysql.user")
	pass := viper.GetString("mysql.password")
	host := viper.GetString("mysql.host")
	port := viper.GetString("mysql.port")
	dbname := viper.GetString("mysql.db")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbname)

	// 🔧 GORM配置优化
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 日志配置：生产环境建议使用 logger.Silent
		Logger: logger.Default.LogMode(logger.Info),

		// 🔧 性能优化
		PrepareStmt:                              true, // 预编译SQL，提升性能
		DisableForeignKeyConstraintWhenMigrating: true, // 迁移时禁用外键

		// 跳过默认事务，需要事务时手动开启
		SkipDefaultTransaction: true,
	})

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// 🔧 获取底层的 sql.DB 并配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get sql.DB: " + err.Error())
	}

	// 连接池配置
	sqlDB.SetMaxOpenConns(100)                 // 最大打开连接数
	sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Hour)        // 连接最大生命周期
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 空闲连接最大生命周期

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		panic("failed to ping database: " + err.Error())
	}

	log.Println("✅ MySQL initialized successfully")
	log.Println("   - Max Open Conns: 100")
	log.Println("   - Max Idle Conns: 10")
	log.Println("   - Conn Max Lifetime: 1h")

	// 自动迁移表结构
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Product{})
	db.AutoMigrate(&model.Order{})
	return db
}
