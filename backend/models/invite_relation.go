package models

import (
	"time"
)

type InviteRelation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	InviterID uint      `gorm:"index" json:"inviter_id"`
	InviteeID uint      `gorm:"index" json:"invitee_id"`
	ShopID    uint      `gorm:"index" json:"shop_id"` // 扫码进入的店铺
	CreatedAt time.Time `json:"created_at"`
}

func (InviteRelation) TableName() string {
	return "invite_relations"
}