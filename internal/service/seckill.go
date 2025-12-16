package service

import (
	"encoding/json"
	"fmt"
	"seckill-system/internal/model"
	mqPkg "seckill-system/internal/pkg/mq"
	redisPkg "seckill-system/internal/pkg/redis"

	"github.com/streadway/amqp"
)

type SeckillService struct{}

func (s *SeckillService) StartSeckill(productID uint, userID uint) error {
	//1.检查用户是否已经购买
	orderKey := fmt.Sprintf("user:product:%d:%d", userID, productID)
	if redisPkg.RDB.Exists(redisPkg.Ctx, orderKey).Val() > 0 {
		return fmt.Errorf("already purchased")
	}

	//2.redis 原子扣减
	key := fmt.Sprintf("stock:%d", productID)
	stock, err := redisPkg.RDB.Decr(redisPkg.Ctx, key).Result()
	if err != nil {
		return err
	}

	//3.库存检查
	if stock < 0 {
		// 库存不足，回滚
		redisPkg.RDB.Incr(redisPkg.Ctx, key)
		return fmt.Errorf("out of stock")
	}

	//4.记录购买信息
	redisPkg.RDB.Set(redisPkg.Ctx, orderKey, 1, 0)

	//5.发送消息到rabbitmq(异步处理)
	message := model.SeckillMessage{
		UserID:    userID,
		ProductID: productID,
	}

	body, err := json.Marshal(message)
	if err != nil {
		redisPkg.RDB.Incr(redisPkg.Ctx, key)
		redisPkg.RDB.Del(redisPkg.Ctx, orderKey)
		return err
	}

	err = mqPkg.Channel.Publish(
		"",              // exchange
		mqPkg.QueueName, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		// 发送消息失败，回滚库存和购买记录
		redisPkg.RDB.Incr(redisPkg.Ctx, key)
		redisPkg.RDB.Del(redisPkg.Ctx, orderKey)
		return err
	}

	return nil //抢购成功
}
