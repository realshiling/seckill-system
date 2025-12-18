package mq

import (
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var (
	Conn      *amqp.Connection // RabbitMQ 连接实例
	Channel   *amqp.Channel    // RabbitMQ 通道实例
	QueueName = "seckill_queue"
)

// 初始化RabbitMQ连接
func InitRabbitMQ() {
	url := viper.GetString("rabbitmq.url")

	// 建立连接（带重试机制）
	var err error
	for i := 0; i < 3; i++ {
		Conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/3): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after 3 attempts: %v", err)
	}

	// 创建通道
	Channel, err = Conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	// 声明队列
	_, err = Channel.QueueDeclare(
		QueueName, // 队列名
		true,      // durable(持久化)
		false,     // autoDelete(自动删除)
		false,     // exclusive(排他)
		false,     // noWait(不等待)
		nil,       // args(参数)
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	log.Println("✅ RabbitMQ initialized successfully")
	log.Println("   - Queue:", QueueName)
	log.Println("   - Durable: true")
}

func Close() {
	if Channel != nil {
		Channel.Close()
	}
	if Conn != nil {
		Conn.Close()
	}
	log.Println("RabbitMQ connections closed")
}
