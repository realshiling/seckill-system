package service

import (
	"encoding/json"
	"fmt"
	"seckill-system/internal/model"
	mqPkg "seckill-system/internal/pkg/mq"
	redisPkg "seckill-system/internal/pkg/redis"
	"time"

	"github.com/streadway/amqp"
)

type SeckillService struct{}

// 限流检查
func (s *SeckillService) RateLimitCheck(userID uint, productID uint) (bool, int, error) {
	//1.构建限流key
	ratelimitkey := fmt.Sprintf("ratelimit:user:%d:product:%d", userID, productID)

	//2.使用redis的incr命令进行限流计数
	count, err := redisPkg.RDB.Incr(redisPkg.Ctx, ratelimitkey).Result()
	if err != nil {
		return false, 0, err
	}

	//3.设置key的过期时间为1分钟
	if count == 1 {
		redisPkg.RDB.Expire(redisPkg.Ctx, ratelimitkey, 1*time.Second)
	}

	//每秒最多一次请求
	maxRequestPerSecond := int64(1)
	remaining := maxRequestPerSecond - count

	if count > maxRequestPerSecond {
		return false, 0, nil
	}

	return true, int(remaining), nil
}

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
