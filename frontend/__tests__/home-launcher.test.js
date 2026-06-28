/**
 * Tests for the home launcher page (Task 4).
 * Home is now a 堂食/外卖 entry launcher, not the menu.
 *   - 堂食: scan a table QR, store the binding, route to the menu page.
 *   - 外卖: navigate straight to the menu in delivery mode; the menu resolves
 *     the delivery shop under its own loading state (no pre-nav silent request).
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

describe('home launcher — 外卖 entry (immediate navigation)', () => {
  it('navigates straight to the menu in delivery mode without a pre-nav resolve', () => {
    pageConfig.chooseDelivery.call({ setData: jest.fn() })
    expect(wx.reLaunch).toHaveBeenCalledWith({
      url: '/pages/menu/index?order_type=delivery',
    })
    // shop resolution is deferred to the menu page; no silent request before navigating
    expect(wx.request).not.toHaveBeenCalled()
  })
})

describe('home launcher — segmented mode', () => {
  it('selectDineIn sets the dine-in segment state', () => {
    const ctx = { setData: jest.fn() }
    pageConfig.selectDineIn.call(ctx)
    expect(ctx.setData).toHaveBeenCalledWith({ mode: 'dine_in' })
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
