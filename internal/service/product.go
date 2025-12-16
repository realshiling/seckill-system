package service

import (
	"fmt"
	"seckill-system/internal/model"
	redisPkg "seckill-system/internal/pkg/redis"

	"gorm.io/gorm"
)

type ProductService struct {
	DB *gorm.DB
}

func (s *ProductService) Create(name string, stock int, price float64) error {
	product := model.Product{
		Name:  name,
		Stock: stock,
		Price: price,
	}
	if err := s.DB.Create(&product).Error; err != nil {
		return err
	}

	//把库存写入redis
	redisKey := fmt.Sprintf("stock:%d", product.ID)
	redisPkg.RDB.Set(redisPkg.Ctx, redisKey, stock, 0)

	return nil
}

func (s *ProductService) List() ([]model.Product, error) {
	var products []model.Product
	err := s.DB.Find(&products).Error
	return products, err
}

// 回灌DB库存到redis
func (s *ProductService) SyncStockToRedis() error {
	var products []model.Product
	if err := s.DB.Select("id", "stock").Find(&products).Error; err != nil {
		return err
	}
	for _, p := range products {
		key := fmt.Sprintf("stock:%d", p.ID)
		exits, err := redisPkg.RDB.Exists(redisPkg.Ctx, key).Result()
		if err != nil {
			return err
		}
		if exits == 0 {
			if err := redisPkg.RDB.Set(redisPkg.Ctx, key, p.Stock, 0).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}
