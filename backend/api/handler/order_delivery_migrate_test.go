package handler

import (
	"testing"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

// OrderDelivery must migrate and persist the delivery detail (address + coords +
// fee + shansong tracking), with a unique constraint on order_id (1:1 with order).
func TestOrderDelivery_MigratesAndEnforcesUniqueOrderID(t *testing.T) {
	setupTestDB(t)

	od := models.OrderDelivery{
		OrderID:         12345,
		RecipientName:   "张三",
		RecipientPhone:  "13800000000",
		Province:        "北京市",
		City:            "北京市",
		County:          "朝阳区",
		DetailAddress:   "测试路 1 号",
		RecipientLat:    39.908722,
		RecipientLng:    116.397499,
		DeliveryFee:     8.50,
		ShansongOrderNo: "SS-TEST-1",
		ShansongStatus:  0,
	}
	if err := config.DB.Create(&od).Error; err != nil {
		t.Fatalf("create order_delivery failed (did migration run?): %v", err)
	}

	var got models.OrderDelivery
	if err := config.DB.Where("order_id = ?", 12345).First(&got).Error; err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if got.DeliveryFee != 8.50 {
		t.Errorf("expected delivery_fee 8.50, got %v", got.DeliveryFee)
	}
	if got.RecipientLat == 0 || got.RecipientLng == 0 {
		t.Errorf("expected recipient coords persisted, got lat=%v lng=%v", got.RecipientLat, got.RecipientLng)
	}

	// order_id is unique — a second delivery row for the same order must be rejected.
	dup := models.OrderDelivery{OrderID: 12345, RecipientName: "李四"}
	if err := config.DB.Create(&dup).Error; err == nil {
		t.Errorf("expected unique-index violation on duplicate order_id, got nil")
	}
}
