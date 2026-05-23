package models

import (
	"time"
)

type Order struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	OrderNo      string    `gorm:"uniqueIndex;size:64" json:"order_no"`
	UserID       uint      `gorm:"index" json:"user_id"`
	ShopID       uint      `gorm:"index" json:"shop_id"`
	TableNo      string    `gorm:"size:32" json:"table_no"`
	Amount       float64   `gorm:"type:numeric(12,2)" json:"amount"`
	RewardAmount float64   `gorm:"type:numeric(12,2);default:0" json:"reward_amount"` // 返利金额
	Status       int       `gorm:"default:1" json:"status"` // 1=pending, 2=paid, 3=completed, 4=cancelled
	PaidAt       *time.Time `json:"paid_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Order) TableName() string {
	return "orders"
}