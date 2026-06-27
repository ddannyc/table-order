/**
 * order-confirm delivery flow (T7): choose address -> get location -> quote;
 * delivery fee enters actual pay; handlePay threads delivery + quote_token;
 * pay blocked until address/quote present.
 */
global.wx = {
  getAccountInfoSync: () => ({ miniProgram: { envVersion: 'develop' } }),
  getStorageSync: jest.fn(() => 'tok'),
  setStorageSync: jest.fn(),
  removeStorageSync: jest.fn(),
  showToast: jest.fn(),
  showModal: jest.fn(),
  chooseAddress: jest.fn(),
  chooseLocation: jest.fn(),
  requestPayment: jest.fn(),
  reLaunch: jest.fn(),
  switchTab: jest.fn(),
}

jest.mock('../api/index.js', () => ({
  getShop: jest.fn(() => Promise.resolve({ reward_ceiling: 0.5 })),
  getRewardBalance: jest.fn(() => Promise.resolve({ reward_balance: 0 })),
  createOrder: jest.fn(() => Promise.resolve({ status: 1, error: 'prepay failed' })),
  getDeliveryQuote: jest.fn(() => Promise.resolve({ delivery_fee: 8.5, quote_token: 'TOK' })),
  getTableBinding: jest.fn(() => ({ shopId: 1, tableNo: '' })),
  setTableBinding: jest.fn(),
}))
jest.mock('../api/product.js', () => ({
  getCart: jest.fn(() => [{ key: '1_0', productId: 1, specId: 0, name: 'Dish', price: 100, quantity: 1 }]),
  clearCart: jest.fn(),
  getCartTotal: jest.fn(() => 100),
}))
jest.mock('../utils/storage.js', () => ({
  doLogin: jest.fn(() => Promise.resolve()),
  handleAuthError: jest.fn(() => false),
  getLastDeliveryAddress: jest.fn(() => null),
  setLastDeliveryAddress: jest.fn(),
}))

let pageConfig
global.Page = (config) => {
  pageConfig = config
}
global.getApp = () => ({})
require('../pages/order-confirm/index.js')
const api = require('../api/index.js')

function makeCtx(dataOverride = {}) {
  const base = {
    shopId: 1, tableNo: '', orderType: 'delivery', orderTypeLabel: '外卖',
    cart: [], cartItems: [], rewardBalance: 0, useReward: false,
    shop: { reward_ceiling: 0.5 }, deliveryFee: '0.00', quoteToken: '',
    deliveryAddress: null, loading: false, needLogin: false,
    totalAmount: '0.00', actualPayAmount: '0.00', rewardDeduct: '0.00', rewardCeiling: '50',
  }
  return {
    setData(patch) { Object.assign(this.data, patch) },
    // sibling methods the page calls on `this`
    applyDeliveryAddress: pageConfig.applyDeliveryAddress,
    fetchQuote: pageConfig.fetchQuote,
    refreshCartDisplay: pageConfig.refreshCartDisplay,
    data: Object.assign(base, dataOverride),
  }
}

beforeEach(() => jest.clearAllMocks())

describe('order-confirm delivery — actual pay includes fee', () => {
  it('adds delivery fee to actual pay amount', () => {
    const ctx = makeCtx({ deliveryFee: '8.50', orderType: 'delivery' })
    pageConfig.refreshCartDisplay.call(ctx)
    expect(ctx.data.totalAmount).toBe('100.00')
    expect(ctx.data.actualPayAmount).toBe('108.50')
  })

  it('does not add a delivery fee for dine_in', () => {
    const ctx = makeCtx({ deliveryFee: '8.50', orderType: 'dine_in' })
    pageConfig.refreshCartDisplay.call(ctx)
    expect(ctx.data.actualPayAmount).toBe('100.00')
  })
})

describe('order-confirm delivery — handlePay', () => {
  it('threads delivery + quote_token into createOrder', () => {
    const ctx = makeCtx({
      orderType: 'delivery', quoteToken: 'TOK', deliveryFee: '8.50',
      totalAmount: '100.00',
      cart: [{ productId: 1, specId: 0, quantity: 1 }],
      deliveryAddress: {
        userName: '张三', telNumber: '13800000000',
        provinceName: '北京市', cityName: '北京市', countyName: '朝阳区',
        detailInfo: '某路1号', lat: 39.9, lng: 116.4,
      },
    })
    pageConfig.handlePay.call(ctx)
    expect(api.createOrder).toHaveBeenCalled()
    const args = api.createOrder.mock.calls[0]
    // (shopId, tableNo, amount, items, useReward, orderType, delivery, quoteToken)
    expect(args[5]).toBe('delivery')
    expect(args[7]).toBe('TOK')
    expect(args[6].recipient_name).toBe('张三')
    expect(args[6].lat).toBe(39.9)
  })

  it('blocks pay when delivery has no address/token', () => {
    const ctx = makeCtx({
      orderType: 'delivery', quoteToken: '', deliveryAddress: null,
      totalAmount: '100.00', cart: [{ productId: 1, specId: 0, quantity: 1 }],
    })
    pageConfig.handlePay.call(ctx)
    expect(api.createOrder).not.toHaveBeenCalled()
    expect(wx.showToast).toHaveBeenCalled()
  })
})

describe('order-confirm delivery — chooseDeliveryAddress', () => {
  it('uses the map-picked point (not device location) for the quote coords', async () => {
    wx.chooseAddress.mockImplementation(({ success }) =>
      success({
        userName: '张三', telNumber: '13800000000',
        provinceName: '北京市', cityName: '北京市', countyName: '朝阳区', detailInfo: '某路1号',
      })
    )
    // Recipient coords come from the chosen delivery point on the map, so a far
    // address quotes against its own location — not wherever the phone is.
    wx.chooseLocation.mockImplementation(({ success }) =>
      success({ latitude: 39.9, longitude: 116.4, name: '收货点', address: '北京市朝阳区某路1号' })
    )

    const ctx = makeCtx({ orderType: 'delivery', shopId: 1 })
    pageConfig.chooseDeliveryAddress.call(ctx)

    expect(api.getDeliveryQuote).toHaveBeenCalled()
    const q = api.getDeliveryQuote.mock.calls[0]
    expect(q[0]).toBe(1)
    expect(q[1].lat).toBe(39.9)

    await Promise.resolve()
    await Promise.resolve()
    expect(ctx.data.deliveryFee).toBe('8.50')
    expect(ctx.data.quoteToken).toBe('TOK')
  })
})

describe('order-confirm wxml — delivery affordances', () => {
  const fs = require('fs')
  const path = require('path')
  const wxml = fs.readFileSync(path.join(__dirname, '../pages/order-confirm/index.wxml'), 'utf8')

  it('binds an address selector to chooseDeliveryAddress', () => {
    expect(wxml).toMatch(/bindtap="chooseDeliveryAddress"/)
  })

  it('renders a delivery fee row', () => {
    expect(wxml).toMatch(/配送费/)
    expect(wxml).toMatch(/\{\{deliveryFee\}\}/)
  })
})
