import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('./client', () => ({
  default: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

import client from './client'
import { getMerchantOrders, prepareOrder, updateOrderStatus, redispatchOrder } from './order'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('order action api', () => {
  it('getMerchantOrders passes params through', () => {
    getMerchantOrders({ shop_id: 5, type: 'delivery' })
    expect(client.get).toHaveBeenCalledWith('/merchant/orders', {
      params: { shop_id: 5, type: 'delivery' },
    })
  })

  it('prepareOrder posts to the prepare endpoint', () => {
    prepareOrder(12)
    expect(client.post).toHaveBeenCalledWith('/merchant/orders/12/prepare')
  })

  it('updateOrderStatus puts the status', () => {
    updateOrderStatus(12, 3)
    expect(client.put).toHaveBeenCalledWith('/merchant/orders/12/status', { status: 3 })
  })

  it('redispatchOrder posts to the redispatch endpoint', () => {
    redispatchOrder(12)
    expect(client.post).toHaveBeenCalledWith('/merchant/orders/12/redispatch')
  })
})
