package models

import (
	"time"
)

type OrderItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OrderID     uint      `gorm:"index" json:"order_id"`
	ProductID   uint      `gorm:"index" json:"product_id"`
	ProductName string    `gorm:"size:128" json:"product_name"`
	Price       float64   `gorm:"type:numeric(12,2)" json:"price"`
	Quantity    int       `json:"quantity"`
	Subtotal    float64   `gorm:"type:numeric(12,2)" json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}