package models

import (
	"time"
)

// OrderDelivery holds the delivery (外卖) detail for a delivery-type order:
// recipient + address + destination coords, the quoted delivery fee, and the
// Shansong dispatch tracking fields. 1:1 with Order (unique order_id).
type OrderDelivery struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	OrderID         uint      `gorm:"uniqueIndex" json:"order_id"`
	RecipientName   string    `gorm:"size:64" json:"recipient_name"`
	RecipientPhone  string    `gorm:"size:32" json:"recipient_phone"`
	Province        string    `gorm:"size:32" json:"province"`
	City            string    `gorm:"size:32" json:"city"`
	County          string    `gorm:"size:32" json:"county"`
	DetailAddress   string    `gorm:"size:255" json:"detail_address"`
	RecipientLat    float64   `gorm:"type:numeric(10,6);default:0" json:"recipient_lat"` // 收件方纬度（wx.getLocation）
	RecipientLng    float64   `gorm:"type:numeric(10,6);default:0" json:"recipient_lng"` // 收件方经度
	DeliveryFee     float64   `gorm:"type:numeric(12,2)" json:"delivery_fee"`
	ShansongQuoteNo string    `gorm:"size:64" json:"shansong_quote_no"`       // 闪送报价凭证，派单时回传
	ShansongOrderNo string    `gorm:"size:64;index" json:"shansong_order_no"` // 闪送返回的运单号
	ShansongStatus  int       `gorm:"default:0" json:"shansong_status"`       // 闪送配送状态码：-1 派单失败 / 0 待派单 / 20 派单中 / 30 待取货 / 40 闪送中 / 50 已完成 / 60 已取消
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (OrderDelivery) TableName() string {
	return "order_deliveries"
}
