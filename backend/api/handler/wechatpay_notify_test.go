package handler

import (
	"testing"
	"time"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

// markOrderPaidOnce must be atomic: only the call that performs the 1->2
// transition wins, so duplicate WeChat notifications can't double-fire reward
// distribution or courier dispatch.
func TestMarkOrderPaidOnce_OnlyFirstWins(t *testing.T) {
	setupTestDB(t)
	user := models.User{OpenID: "pay_user_1"}
	config.DB.Create(&user)
	shop := models.Shop{Name: "Pay Shop"}
	config.DB.Create(&shop)
	order := models.Order{OrderNo: "PAYONCE1", UserID: user.ID, ShopID: shop.ID, OrderType: "dine_in", Amount: 50, Status: 1}
	config.DB.Create(&order)

	now := time.Now()
	first, err := markOrderPaidOnce(order.ID, now)
	if err != nil || !first {
		t.Fatalf("first transition should win: won=%v err=%v", first, err)
	}
	second, err := markOrderPaidOnce(order.ID, now)
	if err != nil {
		t.Fatalf("second transition error: %v", err)
	}
	if second {
		t.Errorf("second transition must not win (order already paid)")
	}

	var got models.Order
	config.DB.First(&got, order.ID)
	if got.Status != 2 || got.PaidAt == nil {
		t.Errorf("order should be paid: status=%d paidAt=%v", got.Status, got.PaidAt)
	}
}
