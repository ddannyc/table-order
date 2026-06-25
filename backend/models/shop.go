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
	Latitude    float64   `gorm:"type:numeric(10,6);default:0" json:"latitude"`  // 门店纬度（外卖寄件方）
	Longitude   float64   `gorm:"type:numeric(10,6);default:0" json:"longitude"` // 门店经度
	RewardRateSelf  float64 `gorm:"type:numeric(5,4);default:0.03" json:"reward_rate_self"`   // 自购返利 3%
	RewardRateLevel1 float64 `gorm:"type:numeric(5,4);default:0.10" json:"reward_rate_level1"` // 直推返利 10%
	RewardRateLevel2 float64 `gorm:"type:numeric(5,4);default:0.04" json:"reward_rate_level2"` // 间推返利 4%
	RewardCeiling float64   `gorm:"type:numeric(5,4);default:0.50" json:"reward_ceiling"`    // 金币抵扣上限 50%
	RewardExcludeCategories string `gorm:"type:jsonb;default:'[]'" json:"reward_exclude_categories"` // 不参与返利的分类
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Shop) TableName() string {
	return "shops"
}