package handler

import (
	"testing"
	"time"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

// T1: Order.PreparedAt (出餐时间) persists as a nullable column — nil by default,
// settable, and round-trips. 已出餐 = PreparedAt != nil.
func TestOrder_PreparedAtPersists(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Prepared Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)

	order := models.Order{OrderNo: "PREP_PERSIST_001", ShopID: shop.ID, Amount: 10, Status: 2}
	config.DB.Create(&order)

	var fresh models.Order
	config.DB.First(&fresh, order.ID)
	if fresh.PreparedAt != nil {
		t.Fatalf("expected PreparedAt nil by default, got %v", fresh.PreparedAt)
	}

	now := time.Now()
	config.DB.Model(&fresh).Update("prepared_at", now)

	var reloaded models.Order
	config.DB.First(&reloaded, order.ID)
	if reloaded.PreparedAt == nil {
		t.Fatalf("expected PreparedAt non-nil after update")
	}
}
