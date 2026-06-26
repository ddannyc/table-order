/**
 * Tests for the menu page interaction logic (Task 5).
 *   - selectCategory: highlights a category and scrolls the list to its anchor.
 *   - switchOrderType: both dine_in and delivery are selectable; delivery
 *     switches to delivery mode and clears any table binding.
 */
global.wx = {
  getAccountInfoSync: jest.fn(() => ({ miniProgram: { envVersion: 'develop' } })),
  getStorageSync: jest.fn(() => ''),
  setStorageSync: jest.fn(),
  removeStorageSync: jest.fn(),
  reLaunch: jest.fn(),
  navigateTo: jest.fn(),
  scanCode: jest.fn(),
  showToast: jest.fn(),
  showModal: jest.fn(),
}

let pageConfig
global.Page = (config) => {
  pageConfig = config
}
global.getApp = () => ({})

require('../pages/menu/index.js')

beforeEach(() => {
  jest.clearAllMocks()
})

describe('menu — selectCategory (weui navbar)', () => {
  it('sets the active category (single panel; no scroll anchor)', () => {
    const ctx = { setData: jest.fn(), data: {} }
    pageConfig.selectCategory.call(ctx, {
      currentTarget: { dataset: { cat: '奶茶牛乳', index: 2 } },
    })
    expect(ctx.setData).toHaveBeenCalledWith({ activeCategory: '奶茶牛乳' })
  })
})

describe('menu — switchOrderType', () => {
  it('selecting 外卖 switches to delivery mode and clears the table binding', () => {
    const ctx = { setData: jest.fn(), data: { orderType: 'dine_in', boundShopId: 1 } }
    pageConfig.switchOrderType.call(ctx, {
      currentTarget: { dataset: { type: 'delivery' } },
    })
    expect(ctx.setData).toHaveBeenCalledWith(
      expect.objectContaining({ orderType: 'delivery', boundTableNo: '' })
    )
    expect(wx.showModal).not.toHaveBeenCalled()
  })

  it('selecting 堂食 sets order type to dine_in', () => {
    const ctx = { setData: jest.fn(), data: { orderType: 'dine_in' } }
    pageConfig.switchOrderType.call(ctx, {
      currentTarget: { dataset: { type: 'dine_in' } },
    })
    expect(ctx.setData).toHaveBeenCalledWith(
      expect.objectContaining({ orderType: 'dine_in' })
    )
    expect(wx.showModal).not.toHaveBeenCalled()
  })
})
