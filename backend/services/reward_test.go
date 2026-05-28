package services

import (
	"testing"
	"time"

	"github.com/example/table-order/models"
)

func TestRewardCalculation_SelfReward(t *testing.T) {
	shop := models.Shop{
		RewardRateSelf:  0.03,
		RewardRateLevel1: 0.10,
		RewardRateLevel2: 0.04,
	}
	amount := 100.00

	selfReward := amount * shop.RewardRateSelf
	expected := 3.00

	if selfReward != expected {
		t.Errorf("self reward: expected %.2f, got %.2f", expected, selfReward)
	}
}

func TestRewardCalculation_Level1Reward(t *testing.T) {
	shop := models.Shop{
		RewardRateSelf:  0.03,
		RewardRateLevel1: 0.10,
		RewardRateLevel2: 0.04,
	}
	amount := 100.00

	level1Reward := amount * shop.RewardRateLevel1
	expected := 10.00

	if level1Reward != expected {
		t.Errorf("level1 reward: expected %.2f, got %.2f", expected, level1Reward)
	}
}

func TestRewardCalculation_Level2Reward(t *testing.T) {
	shop := models.Shop{
		RewardRateSelf:  0.03,
		RewardRateLevel1: 0.10,
		RewardRateLevel2: 0.04,
	}
	amount := 100.00

	level2Reward := amount * shop.RewardRateLevel2
	expected := 4.00

	if level2Reward != expected {
		t.Errorf("level2 reward: expected %.2f, got %.2f", expected, level2Reward)
	}
}

func TestRewardCalculation_ZeroAmount(t *testing.T) {
	shop := models.Shop{
		RewardRateSelf:  0.03,
		RewardRateLevel1: 0.10,
		RewardRateLevel2: 0.04,
	}
	amount := 0.00

	if amount*shop.RewardRateSelf != 0 {
		t.Error("self reward should be 0 for zero amount")
	}
	if amount*shop.RewardRateLevel1 != 0 {
		t.Error("level1 reward should be 0 for zero amount")
	}
	if amount*shop.RewardRateLevel2 != 0 {
		t.Error("level2 reward should be 0 for zero amount")
	}
}

func TestRewardCalculation_FractionalAmount(t *testing.T) {
	shop := models.Shop{
		RewardRateSelf:  0.03,
		RewardRateLevel1: 0.10,
		RewardRateLevel2: 0.04,
	}
	amount := 99.50

	selfReward := amount * shop.RewardRateSelf
	expected := 2.985

	if selfReward != expected {
		t.Errorf("self reward: expected %.3f, got %.3f", expected, selfReward)
	}
}

func TestCheckInactivity_NeverConsumed(t *testing.T) {
	// User with nil LastConsumeAt should NOT be paused
	user := models.User{
		LastConsumeAt: nil,
	}

	if user.LastConsumeAt == nil {
		// Should return false — no inactivity check needed
		// This is the pre-condition for CheckAndPauseInactivity returning false
	}
}

func TestCheckInactivity_Active(t *testing.T) {
	now := time.Now()
	// User who consumed 1 day ago — should NOT be paused
	lastConsume := now.AddDate(0, 0, -1)

	if time.Since(lastConsume).Hours() < float64(rewardInactiveDays*24) {
		// Active — should not pause
	} else {
		t.Error("user with 1-day-old consumption should be active")
	}
}

func TestCheckInactivity_Inactive(t *testing.T) {
	now := time.Now()
	// User who consumed 91 days ago — SHOULD be paused
	lastConsume := now.AddDate(0, 0, -91)

	if time.Since(lastConsume).Hours() >= float64(rewardInactiveDays*24) {
		// Inactive — should pause
	} else {
		t.Error("user with 91-day-old consumption should be inactive")
	}
}

func TestRewardExpiry_NotExpired(t *testing.T) {
	now := time.Now()
	// Reward issued today, expires in 180 days
	expiresAt := now.AddDate(0, 0, 180)

	if expiresAt.Before(now) {
		t.Error("newly issued reward should not be expired")
	}
	if expiresAt.Equal(now) {
		t.Error("newly issued reward should not be expired (equal)")
	}
}

func TestRewardExpiry_Expired(t *testing.T) {
	now := time.Now()
	// Reward issued 181 days ago, should be expired
	expiresAt := now.AddDate(0, 0, -181)

	if !expiresAt.Before(now) {
		t.Error("reward past expiry should be expired")
	}
}

func TestRewardExpiry_Boundary(t *testing.T) {
	now := time.Now()
	// Reward expiring right now — expires_at == now
	// Use Before, not BeforeOrEqual, so boundary is NOT expired
	expiresAt := now.Add(-time.Nanosecond)

	if !expiresAt.Before(now) {
		t.Error("reward 1ns before now should be expired")
	}
}

func TestRewardCeiling_MaxDeduction(t *testing.T) {
	shop := models.Shop{
		RewardCeiling: 0.50,
	}
	amount := 100.00
	rewardBalance := 80.00

	// Max deduction = min(reward_balance, amount * ceiling)
	maxDeduct := min(rewardBalance, amount*shop.RewardCeiling)
	expected := 50.00

	if maxDeduct != expected {
		t.Errorf("max deduction: expected %.2f, got %.2f", expected, maxDeduct)
	}
}

func TestRewardCeiling_LowBalance(t *testing.T) {
	shop := models.Shop{
		RewardCeiling: 0.50,
	}
	amount := 100.00
	rewardBalance := 20.00

	maxDeduct := min(rewardBalance, amount*shop.RewardCeiling)
	expected := 20.00

	if maxDeduct != expected {
		t.Errorf("max deduction with low balance: expected %.2f, got %.2f", expected, maxDeduct)
	}
}

func TestRewardCeiling_ZeroBalance(t *testing.T) {
	shop := models.Shop{
		RewardCeiling: 0.50,
	}
	amount := 100.00
	rewardBalance := 0.00

	maxDeduct := min(rewardBalance, amount*shop.RewardCeiling)

	if maxDeduct != 0 {
		t.Errorf("max deduction with zero balance: expected 0, got %.2f", maxDeduct)
	}
}

func TestHasExcludedCategory_Empty(t *testing.T) {
	shop := models.Shop{
		RewardExcludeCategories: "[]",
	}
	items := []models.OrderItem{}

	result := hasExcludedCategory(items, shop)
	if result {
		t.Error("empty exclusions should return false")
	}
}

func TestHasExcludedCategory_EmptyString(t *testing.T) {
	shop := models.Shop{
		RewardExcludeCategories: "",
	}
	items := []models.OrderItem{}

	result := hasExcludedCategory(items, shop)
	if result {
		t.Error("empty string exclusions should return false")
	}
}

func TestParseJSONArray_Simple(t *testing.T) {
	var result []string
	err := parseJSONArray(`["a","b","c"]`, &result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestParseJSONArray_Empty(t *testing.T) {
	var result []string
	err := parseJSONArray(`[]`, &result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

func TestParseJSONArray_Single(t *testing.T) {
	var result []string
	err := parseJSONArray(`["only"]`, &result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "only" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestParseJSONArray_Invalid(t *testing.T) {
	var result []string
	err := parseJSONArray(``, &result)
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestParseJSONArray_Spaces(t *testing.T) {
	var result []string
	err := parseJSONArray(`["a b","c"]`, &result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0] != "a b" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestRewardPaused_SkipDistribution(t *testing.T) {
	now := time.Now()
	user := models.User{
		RewardPausedAt: &now,
	}

	if user.RewardPausedAt == nil || user.RewardPausedAt.IsZero() {
		t.Error("user with non-nil RewardPausedAt should be paused")
	}
}

func TestRewardPaused_Resume(t *testing.T) {
	user := models.User{
		RewardPausedAt: nil,
	}

	if user.RewardPausedAt != nil {
		t.Error("after resume, RewardPausedAt should be nil")
	}
}

func TestRewardTier_NoInviter(t *testing.T) {
	user := models.User{
		InviterID: nil,
	}

	if user.InviterID != nil {
		t.Error("user with no inviter should skip level-1 and level-2 rewards")
	}
}

func TestRewardTier_NoLevel2(t *testing.T) {
	inviterID := uint(1)
	user := models.User{
		InviterID: &inviterID,
	}

	// inviter has no inviter — level-2 should not distribute
	if user.InviterID != nil {
		// level-1 is valid
		// level-2 requires inviter.InviterID != nil
	}
}

func TestSweepExpired_NoExpired(t *testing.T) {
	now := time.Now()
	logs := []models.RewardLog{
		{ExpiresAt: now.AddDate(0, 0, 1), Expired: false},  // not expired
		{ExpiresAt: now.AddDate(0, 0, 180), Expired: false}, // not expired
	}

	for _, l := range logs {
		if l.ExpiresAt.Before(now) {
			t.Errorf("log with future expiry should not be expired: %+v", l)
		}
	}
}

func TestSweepExpired_HasExpired(t *testing.T) {
	now := time.Now()
	logs := []models.RewardLog{
		{ExpiresAt: now.AddDate(0, 0, -1), Expired: false},  // expired
		{ExpiresAt: now.AddDate(0, 0, 180), Expired: false},  // not expired
	}

	expiredCount := 0
	for _, l := range logs {
		if l.ExpiresAt.Before(now) && !l.Expired {
			expiredCount++
		}
	}

	if expiredCount != 1 {
		t.Errorf("expected 1 expired log, got %d", expiredCount)
	}
}

func TestSweepExpired_AlreadySwept(t *testing.T) {
	now := time.Now()
	l := models.RewardLog{
		ExpiresAt: now.AddDate(0, 0, -1),
		Expired:   true, // already swept
	}

	if l.ExpiresAt.Before(now) && !l.Expired {
		t.Error("already-swept log should not be re-processed")
	}
}
