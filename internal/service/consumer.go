package service

import (
	"encoding/json"
	"fmt"
	"log"
	"seckill-system/internal/model"
	mqPkg "seckill-system/internal/pkg/mq"

	"gorm.io/gorm"
)

type OrderConsumer struct {
	DB *gorm.DB
}

func (oc *OrderConsumer) Start() {
	// 预设消息数量
	err := mqPkg.Channel.Qos(
		10,    // 预设消息数量
		0,     // 大小限制
		false, // 全局
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

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
func (oc *OrderConsumer) handleMessage(body []byte) error {
	var message model.SeckillMessage
	err := json.Unmarshal(body, &message)
	if err != nil {
		return fmt.Errorf("消息解析失败: %v", err)
	}

	//幂等性检查
	log.Printf("📦 [处理中]: UserID=%d, ProductID=%d", message.UserID, message.ProductID)
	var existingorder model.Order
	err = oc.DB.Where("user_id = ? AND product_id = ?", message.UserID, message.ProductID).First(&existingorder).Error

	//订单已存在，返回
	if err == nil {
		log.Printf("⚠️ [已存在订单]: orderID=%d", existingorder.ID)
		return nil
	}

	//没找到订单，创建新订单
	if err != gorm.ErrRecordNotFound {
		// 数据库查询失败
		return fmt.Errorf("查询订单失败: %v", err)
	}

	//事务保证原子性
	err = oc.DB.Transaction(func(tx *gorm.DB) error {
		//1.扣减库存
		result := tx.Model(&model.Product{}).
			Where("id = ? AND stock > 0", message.ProductID).
			Update("stock", gorm.Expr("stock - ?", 1))

		if result.Error != nil {
			return fmt.Errorf("更新库存失败: %v", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("库存不足")
		}

		//2.创建订单
		order := model.Order{
			UserID:    message.UserID,
			ProductID: message.ProductID,
			Status:    "pending",
		}

		err = tx.Create(&order).Error
		if err != nil {
			return fmt.Errorf("订单创建失败: %v", err)
		}

		log.Printf("✅ [订单创建成功]: orderID=%d", order.ID)
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
