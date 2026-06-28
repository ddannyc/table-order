import { describe, it, expect } from 'vitest'
import {
  orderStatusLabel,
  shansongStatusLabel,
  isPaid,
  isPrepared,
  needsAction,
  canPrepare,
  canRedispatch,
  inBucket,
} from './orderBoard'

describe('orderStatusLabel', () => {
  it('maps known statuses', () => {
    expect(orderStatusLabel(1)).toBe('未支付')
    expect(orderStatusLabel(2)).toBe('已支付')
    expect(orderStatusLabel(3)).toBe('已完成')
    expect(orderStatusLabel(4)).toBe('已取消')
  })
  it('falls back for unknown', () => {
    expect(orderStatusLabel(99)).toBe('未知')
  })
})

describe('shansongStatusLabel', () => {
  it('maps the dispatch enum', () => {
    expect(shansongStatusLabel(-1)).toBe('派单失败')
    expect(shansongStatusLabel(0)).toBe('待派单')
    expect(shansongStatusLabel(20)).toBe('派单中')
    expect(shansongStatusLabel(40)).toBe('闪送中')
    expect(shansongStatusLabel(50)).toBe('已完成')
    expect(shansongStatusLabel(60)).toBe('已取消')
  })
})

describe('isPaid / isPrepared', () => {
  it('isPaid is true for status >= 2', () => {
    expect(isPaid({ status: 1 })).toBe(false)
    expect(isPaid({ status: 2 })).toBe(true)
    expect(isPaid({ status: 3 })).toBe(true)
  })
  it('isPrepared reflects prepared_at presence', () => {
    expect(isPrepared({ prepared_at: null })).toBe(false)
    expect(isPrepared({ prepared_at: '2026-06-28T10:00:00Z' })).toBe(true)
  })
})

describe('needsAction (待处理 queue)', () => {
  it('dine_in paid but not prepared needs action', () => {
    expect(needsAction({ order_type: 'dine_in', status: 2, prepared_at: null })).toBe(true)
  })
  it('dine_in already prepared does not', () => {
    expect(needsAction({ order_type: 'dine_in', status: 2, prepared_at: '2026-06-28T10:00:00Z' })).toBe(false)
  })
  it('dine_in unpaid does not (nothing to fulfill yet)', () => {
    expect(needsAction({ order_type: 'dine_in', status: 1, prepared_at: null })).toBe(false)
  })
  it('delivery with failed dispatch needs action', () => {
    expect(needsAction({ order_type: 'delivery', status: 2, delivery: { shansong_status: -1 } })).toBe(true)
  })
  it('delivery dispatching fine does not', () => {
    expect(needsAction({ order_type: 'delivery', status: 2, delivery: { shansong_status: 20 } })).toBe(false)
  })
  it('delivery that was cancelled (60) needs action — it is re-dispatchable', () => {
    expect(needsAction({ order_type: 'delivery', status: 2, delivery: { shansong_status: 60 } })).toBe(true)
  })
})

describe('inBucket (tab partition)', () => {
  const pendingDelivery = { order_type: 'delivery', status: 2, delivery: { shansong_status: -1 } }
  const cancelledDelivery = { order_type: 'delivery', status: 2, delivery: { shansong_status: 60 } }
  const activeDineIn = { order_type: 'dine_in', status: 2, prepared_at: 'x' }
  const doneOrder = { order_type: 'dine_in', status: 3 }
  const cancelledOrder = { order_type: 'dine_in', status: 4 }

  it('pending bucket = needsAction', () => {
    expect(inBucket(pendingDelivery, 'pending')).toBe(true)
    expect(inBucket(cancelledDelivery, 'pending')).toBe(true) // 60 is actionable, must surface here
    expect(inBucket(activeDineIn, 'pending')).toBe(false)
  })
  it('active bucket = paid and not pending', () => {
    expect(inBucket(activeDineIn, 'active')).toBe(true)
    expect(inBucket(pendingDelivery, 'active')).toBe(false)
    expect(inBucket(cancelledDelivery, 'active')).toBe(false) // must NOT hide here
  })
  it('done bucket = completed or cancelled', () => {
    expect(inBucket(doneOrder, 'done')).toBe(true)
    expect(inBucket(cancelledOrder, 'done')).toBe(true)
    expect(inBucket(activeDineIn, 'done')).toBe(false)
  })
  it('all bucket matches everything', () => {
    expect(inBucket(pendingDelivery, 'all')).toBe(true)
    expect(inBucket(doneOrder, 'all')).toBe(true)
  })
})

describe('canPrepare / canRedispatch (action affordances)', () => {
  it('canPrepare only for dine_in paid & unprepared', () => {
    expect(canPrepare({ order_type: 'dine_in', status: 2, prepared_at: null })).toBe(true)
    expect(canPrepare({ order_type: 'dine_in', status: 2, prepared_at: 'x' })).toBe(false)
    expect(canPrepare({ order_type: 'delivery', status: 2, prepared_at: null })).toBe(false)
  })
  it('canRedispatch only for delivery in -1 or 60', () => {
    expect(canRedispatch({ order_type: 'delivery', delivery: { shansong_status: -1 } })).toBe(true)
    expect(canRedispatch({ order_type: 'delivery', delivery: { shansong_status: 60 } })).toBe(true)
    expect(canRedispatch({ order_type: 'delivery', delivery: { shansong_status: 20 } })).toBe(false)
    expect(canRedispatch({ order_type: 'dine_in' })).toBe(false)
  })
})
