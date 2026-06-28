package models

import "time"

// OrderActionLog is an audit trail of merchant-initiated actions on an order
// (出餐 / 重新派单 / 改状态), so manual changes to financial/fulfillment records
// are attributable after the fact.
type OrderActionLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	OrderID    uint      `gorm:"index" json:"order_id"`
	MerchantID uint      `gorm:"index" json:"merchant_id"`
	Action     string    `gorm:"size:32" json:"action"` // prepare | redispatch | status
	OldStatus  int       `json:"old_status"`
	NewStatus  int       `json:"new_status"`
	CreatedAt  time.Time `json:"created_at"`
}

func (OrderActionLog) TableName() string {
	return "order_action_logs"
}
