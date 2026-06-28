// Pure helpers for the order action board. Kept out of the SFC so the triage
// rules are unit-testable and the operational thresholds live in one place.

const ORDER_STATUS_LABELS = { 1: '未支付', 2: '已支付', 3: '已完成', 4: '已取消' }

// Mirrors services.ShansongStatusLabel on the backend.
const SHANSONG_STATUS_LABELS = {
  '-1': '派单失败',
  0: '待派单',
  20: '派单中',
  30: '待取货',
  40: '闪送中',
  50: '已完成',
  60: '已取消',
}

export const orderStatusLabel = (status) => ORDER_STATUS_LABELS[status] || '未知'

export const shansongStatusLabel = (code) => SHANSONG_STATUS_LABELS[code] ?? '未知'

export const isPaid = (order) => order.status >= 2

export const isPrepared = (order) => !!order.prepared_at

// 待处理 queue: an order waiting on the merchant.
//  - dine_in: paid but not yet 出餐
//  - delivery: 闪送 dispatch failed (-1) or cancelled (60) — i.e. re-dispatchable
// (NOTE: 卡太久 time-based stuck detection deferred — thresholds unconfirmed.)
export const needsAction = (order) => {
  if (order.order_type === 'delivery') {
    return canRedispatch(order)
  }
  return isPaid(order) && !isPrepared(order)
}

// Tab partition for the board. Kept here (not in the SFC) so it is unit-tested
// and stays consistent with needsAction.
export const inBucket = (order, which) => {
  switch (which) {
    case 'pending':
      return needsAction(order)
    case 'active':
      return order.status === 2 && !needsAction(order)
    case 'done':
      return order.status === 3 || order.status === 4
    default:
      return true
  }
}

export const canPrepare = (order) =>
  order.order_type !== 'delivery' && isPaid(order) && !isPrepared(order)

// Only a paid (status 2) delivery order is re-dispatchable — a merchant-cancelled
// order must not show the button (mirrors the backend order.Status==2 guard).
export const canRedispatch = (order) =>
  order.order_type === 'delivery' &&
  order.status === 2 &&
  (order.delivery?.shansong_status === -1 || order.delivery?.shansong_status === 60)
