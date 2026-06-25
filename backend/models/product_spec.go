package models

import (
	"time"
)

// ProductSpec is a selectable variant of a product (e.g. 600ml / 800ml),
// each with its own price. Products without specs are treated as a single
// default spec priced at the product's own price.
type ProductSpec struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"index" json:"product_id"`
	Name      string    `gorm:"size:64" json:"name"`
	Price     float64   `gorm:"type:numeric(12,2)" json:"price"`
	Status    int       `gorm:"default:1" json:"status"` // 1=上架, 0=下架, 2=售罄
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ProductSpec) TableName() string {
	return "product_specs"
}
