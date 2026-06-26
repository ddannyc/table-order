/**
 * Tests for the home launcher page (Task 4).
 * Home is now a 堂食/外卖 entry launcher, not the menu.
 *   - 堂食: scan a table QR, store the binding, route to the menu page.
 *   - 外卖: resolve the delivery shop, route to the menu in delivery mode.
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

describe('home launcher — 外卖 entry (enabled)', () => {
  it('resolves the delivery shop and routes to the menu in delivery mode', async () => {
    wx.request.mockImplementation(({ success }) =>
      success({ statusCode: 200, data: { id: 7, name: 'Shop' } })
    )
    await pageConfig.chooseDelivery()
    expect(wx.reLaunch).toHaveBeenCalledWith({
      url: '/pages/menu/index?order_type=delivery&shop_id=7',
    })
  })

  it('shows a hint and does not navigate when no delivery shop is available', async () => {
    wx.request.mockImplementation(({ success }) =>
      success({ statusCode: 404, data: { error: 'no available shop' } })
    )
    await pageConfig.chooseDelivery()
    expect(wx.reLaunch).not.toHaveBeenCalled()
    const notified = wx.showModal.mock.calls.length + wx.showToast.mock.calls.length
    expect(notified).toBeGreaterThan(0)
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

describe('home — wallet header (R2)', () => {
  it('loads real balance and reward when logged in', async () => {
    wx.getStorageSync.mockImplementation((k) => (k === 'token' ? 'tok' : ''))
    wx.request.mockImplementation(({ url, success }) => {
      if (url.includes('/wallet/balance')) {
        success({ statusCode: 200, data: { balance: 12.5, reward_balance: 3 } })
      } else if (url.includes('/reward/balance')) {
        success({ statusCode: 200, data: { reward_balance: 3, reward_paused: false } })
      } else {
        success({ statusCode: 200, data: {} })
      }
    })
    const ctx = { setData: jest.fn(), data: {} }
    await pageConfig.loadWallet.call(ctx)
    const merged = Object.assign({}, ...ctx.setData.mock.calls.map((c) => c[0]))
    expect(merged.balanceText).toBe('12.50')
    expect(merged.rewardText).toBe('3.00')
  })

  it('degrades to a dash when not logged in (no fake numbers)', async () => {
    wx.getStorageSync.mockImplementation(() => '')
    const ctx = { setData: jest.fn(), data: {} }
    await pageConfig.loadWallet.call(ctx)
    const merged = Object.assign({}, ...ctx.setData.mock.calls.map((c) => c[0]))
    expect(merged.balanceText).toBe('—')
    expect(merged.rewardText).toBe('—')
    // never hit the wallet endpoint without a token
    expect(wx.request).not.toHaveBeenCalled()
  })
})
