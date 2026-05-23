package models

import (
	"time"
)

type WalletLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Type      string    `gorm:"size:16" json:"type"` // reward, deduction, invite_reward
	Amount    float64   `gorm:"type:numeric(12,2)" json:"amount"`
	OrderID   *uint     `gorm:"index" json:"order_id"`
	Remark    string    `gorm:"size:256" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}

func (WalletLog) TableName() string {
	return "wallet_logs"
}