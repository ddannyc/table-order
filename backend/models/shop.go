package models

import (
	"time"
)

type Shop struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	MerchantID  uint      `gorm:"index" json:"merchant_id"`
	Name        string    `gorm:"size:128" json:"name"`
	Description string    `gorm:"size:512" json:"description"`
	Address     string    `gorm:"size:256" json:"address"`
	Phone       string    `gorm:"size:32" json:"phone"`
	Hours       string    `gorm:"size:128" json:"hours"`
	Logo        string    `gorm:"size:512" json:"logo"`
	RewardRate     float64   `gorm:"type:numeric(5,4);default:0.1" json:"reward_rate"`     // 返利比例，默认10%
	InviteRate    float64   `gorm:"type:numeric(5,4);default:0.05" json:"invite_rate"`    // 邀请奖励比例，默认5%
	RewardCeiling float64   `gorm:"type:numeric(5,4);default:0.8" json:"reward_ceiling"`  // 福利金抵扣上限比例，默认80%
	Status      int       `gorm:"default:1" json:"status"` // 1=active, 0=inactive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Shop) TableName() string {
	return "shops"
}