package services

import (
	"context"
	"log"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

// shansongDispatchedStatus is the initial status after a successful dispatch
// (派单中, per the merchants/v5 status enum). Later statuses arrive via callback.
const shansongDispatchedStatus = 20

// shansongFailedStatus marks a paid delivery order whose orderPlace call failed,
// so the stranded order is queryable (and shown as 派单失败) instead of being
// silently masked as 配送中. Recovery (retry/re-dispatch) is operator-driven.
const shansongFailedStatus = -1

// DispatchShansong places the Shansong courier order for a paid delivery order.
// Async, best-effort: any failure is logged and never blocks the payment flow.
// No-op for non-delivery orders (no OrderDelivery row) and idempotent once a
// Shansong order number is already recorded.
func DispatchShansong(orderID uint) {
	var od models.OrderDelivery
	if err := config.DB.Where("order_id = ?", orderID).First(&od).Error; err != nil {
		return // not a delivery order (or no detail) — nothing to dispatch
	}
	if od.ShansongOrderNo != "" {
		return // already dispatched — idempotent against duplicate pay notifications
	}
	if Shansong == nil {
		log.Printf("[shansong] client not configured, skip dispatch orderID=%d", orderID)
		return
	}

	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		log.Printf("[shansong] order not found orderID=%d: %v", orderID, err)
		return
	}

	// orderPlace confirms the prior quote — only the issOrderNo is needed.
	no, err := Shansong.CreateOrder(context.Background(), CreateDeliveryRequest{
		QuoteToken: od.ShansongQuoteNo,
		OrderNo:    order.OrderNo,
	})
	if err != nil {
		log.Printf("[shansong] dispatch failed orderID=%d: %v", orderID, err)
		// Surface the failure durably so the paid-but-undispatched order can be
		// found and re-dispatched, rather than living only in the log.
		if uerr := config.DB.Model(&od).Update("shansong_status", shansongFailedStatus).Error; uerr != nil {
			log.Printf("[shansong] persist dispatch-failure status failed orderID=%d: %v", orderID, uerr)
		}
		return
	}

	if err := config.DB.Model(&od).Updates(map[string]any{
		"shansong_order_no": no,
		"shansong_status":   shansongDispatchedStatus,
	}).Error; err != nil {
		log.Printf("[shansong] persist dispatch result failed orderID=%d shansongNo=%s: %v", orderID, no, err)
	}
	log.Printf("[shansong] dispatched orderID=%d shansongNo=%s", orderID, no)
}
