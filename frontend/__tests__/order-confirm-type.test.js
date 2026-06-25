/**
 * Tests that order-confirm reflects order_type from the route (Task 7).
 * order-confirm is read-only on type — it shows what home/menu chose.
 */
global.wx = {
  getAccountInfoSync: () => ({ miniProgram: { envVersion: 'develop' } }),
  getStorageSync: jest.fn(() => ''),
  setStorageSync: jest.fn(),
  removeStorageSync: jest.fn(),
  showModal: jest.fn(),
  switchTab: jest.fn(),
}

let pageConfig
global.Page = (config) => {
  pageConfig = config
}
global.getApp = () => ({})

require('../pages/order-confirm/index.js')

beforeEach(() => {
  jest.clearAllMocks()
})

describe('order-confirm — order_type from route', () => {
  it('sets orderType and a label from the delivery route option', () => {
    const ctx = {
      setData: jest.fn(),
      refreshCartDisplay: jest.fn(),
      checkAuthAndLoad: jest.fn(),
      data: {},
    }
    pageConfig.onLoad.call(ctx, { shop_id: '1', table_no: '', order_type: 'delivery' })
    expect(ctx.setData).toHaveBeenCalledWith(
      expect.objectContaining({ orderType: 'delivery', orderTypeLabel: '外卖' })
    )
  })

  it('defaults to dine_in label when no order_type in route', () => {
    const ctx = {
      setData: jest.fn(),
      refreshCartDisplay: jest.fn(),
      checkAuthAndLoad: jest.fn(),
      data: {},
    }
    pageConfig.onLoad.call(ctx, { shop_id: '1', table_no: 'A01' })
    const merged = Object.assign({}, ...ctx.setData.mock.calls.map((c) => c[0]))
    expect(merged.orderType).toBe('dine_in')
    expect(merged.orderTypeLabel).toBe('堂食')
  })
})
