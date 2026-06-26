/**
 * Tests for the menu page interaction logic.
 *   - selectCategory: highlights a category (single-panel filter).
 *   - order type at entry: no in-menu switch; mode is derived from how the
 *     menu was entered (delivery flag vs table binding), and a delivery entry
 *     clears any stale dine-in table binding.
 */
const fs = require('fs')
const path = require('path')
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

describe('menu — order type at entry (no in-menu switch)', () => {
  it('no longer exposes an in-menu order-type switch', () => {
    expect(pageConfig.switchOrderType).toBeUndefined()
  })

  it('entering via the delivery flag sets delivery mode and clears any stale table binding', () => {
    const ctx = { setData: jest.fn(), loadData: jest.fn() }
    pageConfig.onLoad.call(ctx, { order_type: 'delivery', shop_id: '7' })
    expect(ctx.setData).toHaveBeenCalledWith(
      expect.objectContaining({ orderType: 'delivery', boundTableNo: '', boundShopId: 7 })
    )
    expect(wx.removeStorageSync).toHaveBeenCalledWith('current_table_no')
    expect(ctx.loadData).toHaveBeenCalled()
  })

  it('entering with a table binding stays in dine-in mode', () => {
    const ctx = { setData: jest.fn(), loadData: jest.fn() }
    pageConfig.onLoad.call(ctx, { shop_id: '3', table_no: 'A01' })
    expect(wx.setStorageSync).toHaveBeenCalledWith('current_shop_id', 3)
    expect(ctx.setData).toHaveBeenCalledWith(
      expect.objectContaining({ boundShopId: 3, boundTableNo: 'A01' })
    )
    // dine-in is the default; no delivery mode set
    const setDeliveryCalls = ctx.setData.mock.calls.filter(
      ([arg]) => arg && arg.orderType === 'delivery'
    )
    expect(setDeliveryCalls).toHaveLength(0)
  })

  it('renders a read-only mode indicator, not a toggle', () => {
    const wxml = fs.readFileSync(
      path.join(__dirname, '../pages/menu/index.wxml'),
      'utf8'
    )
    expect(wxml).not.toMatch(/bindtap="switchOrderType"/)
    expect(wxml).not.toMatch(/menu-typetoggle/)
  })
})

describe('menu — left-right layout + big image cards (M4)', () => {
  it('falls back to the category placeholder when a card image errors', () => {
    const setData = jest.fn()
    const ctx = {
      setData,
      data: {
        activeCategory: '奶茶牛乳',
        productsByCategory: { 奶茶牛乳: [{ id: 9, hasImage: true }] },
      },
    }
    pageConfig.onImgError.call(ctx, { currentTarget: { dataset: { id: 9 } } })
    expect(setData).toHaveBeenCalledWith({
      productsByCategory: { 奶茶牛乳: [{ id: 9, hasImage: false }] },
    })
  })

  it('uses a left category rail and big image cards (not the old horizontal navbar)', () => {
    const wxml = fs.readFileSync(
      path.join(__dirname, '../pages/menu/index.wxml'),
      'utf8'
    )
    expect(wxml).toMatch(/menu-rail/)
    expect(wxml).toMatch(/menu-card/)
    expect(wxml).toMatch(/menu-thumb-ph/) // CSS placeholder block
    expect(wxml).not.toMatch(/menu-navbar/) // old horizontal navbar removed
  })
})
