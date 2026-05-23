package models

import (
	"time"
)

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ShopID      uint      `gorm:"index" json:"shop_id"`
	Name        string    `gorm:"size:128" json:"name"`
	Price       float64   `gorm:"type:numeric(12,2)" json:"price"`
	Description string    `gorm:"size:512" json:"description"`
	Image       string    `gorm:"size:512" json:"image"`
	Category    string    `gorm:"size:64" json:"category"`
	Status      int       `gorm:"default:1" json:"status"` // 1=上架, 0=下架, 2=售罄
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Product) TableName() string {
	return "products"
}