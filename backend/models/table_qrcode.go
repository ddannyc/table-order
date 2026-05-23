package models

import (
	"time"
)

type TableQRCode struct {
	ID      uint      `gorm:"primaryKey" json:"id"`
	ShopID  uint      `gorm:"index" json:"shop_id"`
	TableNo string    `gorm:"size:32" json:"table_no"`
	QRCode  string    `gorm:"size:512" json:"qrcode"`
	Token   string    `gorm:"uniqueIndex;size:64" json:"token"` // 用于扫码跳转的随机token
	Status  int       `gorm:"default:1" json:"status"` // 1=active, 0=inactive
	CreatedAt time.Time `json:"created_at"`
}

func (TableQRCode) TableName() string {
	return "table_qrcodes"
}