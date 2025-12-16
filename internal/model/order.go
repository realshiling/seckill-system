package model

import "time"

type Order struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;index"`    // 用户ID
	ProductID uint   `gorm:"not null;index"`    // 商品ID
	Status    string `gorm:"default:'pending'"` // pending, paid, cancelled
	CreatedAt time.Time
	UpdatedAt time.Time
}
