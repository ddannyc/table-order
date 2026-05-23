package models

import (
	"time"
)

type User struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	OpenID        string     `gorm:"uniqueIndex;size:128" json:"openid"`
	UnionID       string     `gorm:"index;size:128" json:"unionid"`
	Nickname      string     `gorm:"size:64" json:"nickname"`
	Avatar        string     `gorm:"size:512" json:"avatar"`
	Phone         string     `gorm:"index;size:32" json:"phone"`
	InviterID     *uint      `gorm:"index" json:"inviter_id"`
	Balance       float64    `gorm:"type:numeric(12,2);default:0" json:"balance"`
	RewardBalance float64    `gorm:"type:numeric(12,2);default:0" json:"reward_balance"`
	Role          int        `gorm:"default:0" json:"role"`
	IsBanned      bool       `gorm:"default:false" json:"is_banned"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}