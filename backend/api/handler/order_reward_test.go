package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"gorm.io/gorm/clause"
)

// The reward deduction is a guarded conditional update: if the balance is
// drained by a concurrent order between the handler's read and its write, the
// deduction must NOT apply (no double-spend, balance never goes negative).
//
// This is forced deterministically: we hold the user row under SELECT ... FOR
// UPDATE so the handler parks on its deduction UPDATE, then drain the balance
// and commit. When the handler unblocks, its `reward_balance >= deduct` guard
// must reject the write and roll the whole order back.
func TestCreateOrder_RewardDeductionGuardedAgainstConcurrentDrain(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Drain Shop", MerchantID: 1, Status: 1, RewardCeiling: 1.0}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Drain Dish", Price: 100, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "drain_reward_user", Nickname: "Drain", Role: 0, RewardBalance: 100, Balance: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":    shop.ID,
		"table_no":   "A01",
		"amount":     100,
		"use_reward": true,
		"items":      []map[string]interface{}{{"product_id": product.ID, "quantity": 1}},
	}
	jsonBody, _ := json.Marshal(body)

	// Hold the user row so the handler's deduction UPDATE will block.
	lockTx := config.DB.Begin()
	var locked models.User
	if err := lockTx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, user.ID).Error; err != nil {
		lockTx.Rollback()
		t.Fatalf("failed to lock user row: %v", err)
	}

	type result struct {
		code int
		body string
	}
	done := make(chan result, 1)
	go func() {
		req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		done <- result{w.Code, w.Body.String()}
	}()

	// Let the handler reach (and block on) its guarded deduction UPDATE, then
	// drain the balance and commit — simulating a concurrent order that spent it.
	time.Sleep(300 * time.Millisecond)
	if err := lockTx.Model(&models.User{}).Where("id = ?", user.ID).Update("reward_balance", 0).Error; err != nil {
		lockTx.Rollback()
		t.Fatalf("failed to drain balance: %v", err)
	}
	lockTx.Commit()

	res := <-done

	if res.code != http.StatusBadRequest {
		t.Errorf("expected 400 (deduction rejected after drain), got %d", res.code)
	}
	// Assert it failed for the right reason — the guard, not some unrelated 400.
	if !strings.Contains(res.body, "reward balance insufficient") {
		t.Errorf("expected 'reward balance insufficient' rejection, got body: %s", res.body)
	}
	config.DB.First(&user, user.ID)
	if user.RewardBalance < 0 {
		t.Fatalf("reward balance went negative: %v (double-spend)", user.RewardBalance)
	}
	if user.RewardBalance != 0 {
		t.Errorf("expected reward_balance to stay 0 (drained, no further deduction), got %v", user.RewardBalance)
	}
	var orderCount int64
	config.DB.Model(&models.Order{}).Where("user_id = ?", user.ID).Count(&orderCount)
	if orderCount != 0 {
		t.Errorf("expected the order to roll back (0 persisted), got %d", orderCount)
	}
	// The whole tx rolled back, so no deduct wallet log should exist.
	var deductLogs int64
	config.DB.Model(&models.WalletLog{}).Where("user_id = ? AND type = ?", user.ID, "deduct").Count(&deductLogs)
	if deductLogs != 0 {
		t.Errorf("expected no deduct wallet log after rollback, got %d", deductLogs)
	}
}
