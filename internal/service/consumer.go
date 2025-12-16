package service

import (
	"encoding/json"
	"log"
	"seckill-system/internal/model"
	mqPkg "seckill-system/internal/pkg/mq"

	"gorm.io/gorm"
)

type OrderConsumer struct {
	DB *gorm.DB
}

func (oc *OrderConsumer) Start() {
	// 启动订单消费逻辑
	msgs, err := mqPkg.Channel.Consume(
		mqPkg.QueueName, // queue
		"",              // consumer
		false,           // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}
	log.Println("Order consumer started,waiting for messages...")

	// 启动一个 goroutine 来处理消息
	go func() {
		for msg := range msgs {
			oc.handleMessage(msg.Body)
			msg.Ack(false)
		}
	}()
}

// 处理消息的具体逻辑
func (oc *OrderConsumer) handleMessage(body []byte) {
	var message model.SeckillMessage
	err := json.Unmarshal(body, &message)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	log.Printf("processing order: UserID=%d, ProductID=%d", message.UserID, message.ProductID)

	//1.更新数据库库存
	err = oc.DB.Model(&model.Product{}).
		Where("id = ?", message.ProductID).
		UpdateColumn("stock", gorm.Expr("stock - ?", 1)).
		Error

	if err != nil {
		log.Printf("Failed to update stock in DB: %v", err)
		return
	}

	//2.创建订单
	order := model.Order{
		UserID:    message.UserID,
		ProductID: message.ProductID,
		Status:    "pending",
	}

	err = oc.DB.Create(&order).Error
	if err != nil {
		log.Printf("Failed to create order in DB: %v", err)
		return
	}

	log.Printf("Order created successfully: OrderID=%d", order.ID)
}
