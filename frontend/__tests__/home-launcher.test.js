/**
 * Tests for the home launcher page (Task 4).
 * Home is now a 堂食/外卖 entry launcher, not the menu.
 *   - 堂食: scan a table QR, store the binding, route to the menu page.
 *   - 外卖: placeholder until Phase 4 — must not enter an unfinished flow.
 */
global.wx = {
  getAccountInfoSync: jest.fn(() => ({ miniProgram: { envVersion: 'develop' } })),
  getStorageSync: jest.fn(() => ''),
  setStorageSync: jest.fn(),
  removeStorageSync: jest.fn(),
  reLaunch: jest.fn(),
  navigateTo: jest.fn(),
  switchTab: jest.fn(),
  scanCode: jest.fn(),
  chooseAddress: jest.fn(),
  request: jest.fn(),
  showToast: jest.fn(),
  showModal: jest.fn(),
}

let pageConfig
global.Page = (config) => {
  pageConfig = config
}
global.getApp = () => ({})

require('../pages/home/index.js')

beforeEach(() => {
  jest.clearAllMocks()
})

describe('home launcher — 外卖 entry (chooseAddress → delivery menu)', () => {
  it('caches the chosen address and routes to the delivery menu with the resolved shop', async () => {
    wx.chooseAddress.mockImplementation(({ success }) =>
      success({ userName: '张三', telNumber: '13800000000', provinceName: '上海', detailInfo: '世纪广场' })
    )
    // resolveDeliveryShop() -> request() -> wx.request success with the shop
    wx.request.mockImplementation(({ success }) => success({ statusCode: 200, data: { id: 7 } }))

    pageConfig.chooseDelivery()
    await new Promise((r) => setTimeout(r, 0)) // flush the resolve promise

    expect(wx.setStorageSync).toHaveBeenCalledWith(
      'last_delivery_address',
      expect.objectContaining({ userName: '张三' })
    )
    expect(wx.reLaunch).toHaveBeenCalledWith({
      url: '/pages/menu/index?order_type=delivery&shop_id=7',
    })
  })
})

describe('home launcher — 堂食 entry (scan to menu)', () => {
  it('scans a table QR, stores the binding, and routes to the menu page', () => {
    wx.scanCode.mockImplementation(({ success }) =>
      success({ result: 'https://x/scan?shop_id=1&table_no=A01' })
    )
    pageConfig.scanDineIn()
    expect(wx.setStorageSync).toHaveBeenCalledWith('current_shop_id', 1)
    expect(wx.setStorageSync).toHaveBeenCalledWith('current_table_no', 'A01')
    expect(wx.reLaunch).toHaveBeenCalledWith({ url: '/pages/menu/index' })
  })
})
