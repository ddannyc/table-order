package models

import (
	"time"
)

type Merchant struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Phone     string    `gorm:"uniqueIndex;size:32" json:"phone"`
	Password  string    `gorm:"size:256" json:"-"`
	Name      string    `gorm:"size:64" json:"name"`
	Company   string    `gorm:"size:128" json:"company"`
	Status    int       `gorm:"default:1" json:"status"` // 1=active, 0=inactive
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Merchant) TableName() string {
	return "merchants"
}