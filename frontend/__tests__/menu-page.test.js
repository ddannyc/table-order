/**
 * Tests for the menu page interaction logic (Task 5).
 *   - selectCategory: highlights a category and scrolls the list to its anchor.
 *   - switchOrderType: dine_in selectable now; delivery shows a coming-soon
 *     placeholder and does NOT switch (Phase 4 wires delivery).
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

describe('menu — selectCategory', () => {
  it('sets the active category and scroll-into-view anchor by index', () => {
    const ctx = { setData: jest.fn(), data: {} }
    pageConfig.selectCategory.call(ctx, {
      currentTarget: { dataset: { cat: '奶茶牛乳', index: 2 } },
    })
    expect(ctx.setData).toHaveBeenCalledWith(
      expect.objectContaining({ activeCategory: '奶茶牛乳', scrollIntoId: 'cat-2' })
    )
  })
})

describe('menu — switchOrderType', () => {
  it('selecting 外卖 shows coming-soon and does not switch order type', () => {
    const ctx = { setData: jest.fn(), data: { orderType: 'dine_in' } }
    pageConfig.switchOrderType.call(ctx, {
      currentTarget: { dataset: { type: 'delivery' } },
    })
    expect(wx.showModal).toHaveBeenCalled()
    expect(ctx.setData).not.toHaveBeenCalledWith(
      expect.objectContaining({ orderType: 'delivery' })
    )
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
