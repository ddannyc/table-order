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

describe('home launcher — 外卖 entry (coming soon, gated)', () => {
  it('shows a coming-soon hint and does not enter the delivery flow', () => {
    pageConfig.chooseDelivery()
    const notified = wx.showModal.mock.calls.length + wx.showToast.mock.calls.length
    expect(notified).toBeGreaterThan(0)
    expect(wx.reLaunch).not.toHaveBeenCalled()
    expect(wx.navigateTo).not.toHaveBeenCalled()
    expect(wx.chooseAddress).not.toHaveBeenCalled()
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
