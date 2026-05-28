package models

import (
	"time"
)

type RewardLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	OrderID   uint      `gorm:"index;not null" json:"order_id"`
	Type      string    `gorm:"size:16;not null" json:"type"` // self, invite_level1, invite_level2
	Amount    float64   `gorm:"type:numeric(12,2);not null" json:"amount"`
	FromUserID *uint    `gorm:"index" json:"from_user_id"` // 下单人 (来源)
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Expired   bool      `gorm:"default:false;index" json:"expired"`
	CreatedAt time.Time `json:"created_at"`
}

func (RewardLog) TableName() string {
	return "reward_logs"
}
