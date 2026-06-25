/**
 * Tests for createOrder order_type threading (Task 7).
 * Source of truth for order_type is the home/menu choice; it must reach the
 * /orders payload. Defaults to dine_in for backward compatibility.
 */
global.wx = {
  getAccountInfoSync: () => ({ miniProgram: { envVersion: 'develop' } }),
  getStorageSync: jest.fn(() => 'tok'),
  removeStorageSync: jest.fn(),
  request: jest.fn(),
}

const { createOrder } = require('../api/index.js')

beforeEach(() => {
  jest.clearAllMocks()
})

describe('createOrder — order_type', () => {
  it('defaults order_type to dine_in when not provided', () => {
    createOrder(1, 'A01', 100, [], false)
    const data = wx.request.mock.calls[0][0].data
    expect(data.order_type).toBe('dine_in')
    expect(data.table_no).toBe('A01')
  })

  it('sends delivery order_type when provided', () => {
    createOrder(1, '', 100, [], false, 'delivery')
    const data = wx.request.mock.calls[0][0].data
    expect(data.order_type).toBe('delivery')
    expect(data.table_no).toBe('')
  })
})
