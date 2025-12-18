# Seckill System（秒杀系统）

## 简介
- 基于 Go + Gin 的秒杀示例，含 Redis 预扣减、RabbitMQ 异步下单、MySQL 持久化。
- 简化的分层：handler（HTTP）→ service（业务）→ DB/缓存/MQ。

## 核心流程
1. HTTP 请求发起秒杀。
2. Redis 限流 + 去重：同一用户/商品每秒限 1 次，已购直接拒绝。
3. Redis 原子扣减库存（`DECR`），库存不足回滚。
4. 写入“已购”标记到 Redis。
5. 推送秒杀消息到 RabbitMQ。
6. 消费者（OrderConsumer）从 MQ 拉取消息，在 MySQL 里事务扣库存 + 写订单（两步要么都成功，要么都回滚）。

## 依赖
- Go 1.21+
- MySQL
- Redis
- RabbitMQ

## 配置
在 `internal/config/config.yaml`（名称/路径按你项目调整）填写连接信息：
```yaml
mysql:
  dsn: "user:pass@tcp(127.0.0.1:3306)/seckill?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  poolSize: 20

rabbitmq:
  url: "amqp://user:pass@127.0.0.1:5672/"
```

## 快速开始
```bash
# 1) 拉取依赖
go mod tidy

# 2) 确保本机已启动 MySQL、Redis、RabbitMQ 并配置好 config.yaml

# 3) 运行 API 与消费者（订单消费者在 main 中自动启动）
go run ./cmd/api/main.go
```

## 主要接口（示例）
- POST `/register` 用户注册
- POST `/login` 用户登录
- （可在 handler 中扩展商品列表、创建商品、发起秒杀等接口）

## 目录速览
```
cmd/api/main.go        # 入口，初始化 DB/Redis/RabbitMQ，启动消费者与 HTTP
internal/handler       # HTTP 层（Gin）
internal/service       # 业务逻辑（商品、秒杀、订单消费者等）
internal/model         # 数据模型
internal/pkg/redis     # Redis 客户端
internal/pkg/mq        # RabbitMQ 客户端
internal/database      # MySQL 初始化（GORM）
```

## 运行时要点
- 连接池：GORM 与 go-redis 默认自带连接池，可在初始化时配置最大连接数/池大小。
- 消费者：OrderConsumer 使用事务保证 MySQL 扣库存与创建订单的原子性。
- Redis 预扣减能抗高并发；若 MQ 发送失败，会回滚库存与购买标记。
- 建议为 `orders(user_id, product_id)` 建唯一索引，进一步防重复下单。

## 常用命令
```bash
git status
git add .
git commit -m "msg"
git push origin main
```

## 学习收获
1.GORM 与 事务
    问题：
        tx.Model(&model.Product{}) 到底如何定位表
        事务如何保证“扣库存+下单”原子性
    收获：  
        1）GORM 会根据传入的结构体类型定位表名
        2）tx.Transaction 方法中tx一切对数据库的操作都在同一个事务中

2.指针，依赖注入与分层
    问题：
        handler 包 service，service 包 DB 的关系
    收获：
        1）handler 负责 HTTP/路由，
        2）service 承载业务，向下用 DB/Redis/MQ；
        3）main 中把依赖通过指针注入

3.redis 与消息队列
    问题：
        写入redis库存前需要键名拼接
        redisPkg用途
    收获：
        1）redis库存键名拼接，防止冲突
        2）redisPkg 封装 redis 客户端，方便全局使用
        3）redisCtx 传递超时/取消/链路信息
        4）Publish 使用默认交换机，routing key 为队列名，消息入对应队列；
            消费者从队列取消息后在 MySQL 里做事务扣减+下单

4.连接池与运行时
    困难：
        连接池是否需要显式调用；
        连接池连接了什么
    知识点：
        1）GORM 内部 *sql.DB、go-redis 都自带连接池；
            初始化时配置一次（MaxOpen/MaxIdle/PoolSize 等）即可
        2）连接池是“应用进程 ↔ 服务端（MySQL/Redis）”的连接集合；
            RabbitMQ 通常是单连接多 channel，无池