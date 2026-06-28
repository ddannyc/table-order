/**
 * 外卖冷启动落在菜单页：首页点「外卖」立即跳转、不在导航前解析门店。
 * 菜单页负责：无 shop_id 时自己 resolveDeliveryShop → loadData(门店)，
 * 并复用该门店 DTO 跳过冗余的 getShop（/delivery/shop 与 getShop 同为 toPublicShopDTO）。
 * 无可配送门店时由菜单页给出 error 态 + toast；onRetry 在未绑定门店时重走解析。
 */
jest.mock('../api/index.js', () => ({
  getShop: jest.fn(),
  getTableBinding: jest.fn(() => ({ shopId: 0, tableNo: '' })),
  setTableBinding: jest.fn(),
  clearTableBinding: jest.fn(),
  bindInviteCode: jest.fn(() => Promise.resolve()),
  resolveDeliveryShop: jest.fn(),
}))
jest.mock('../api/product.js', () => ({
  getShopProducts: jest.fn(() => Promise.resolve([])),
  getCart: jest.fn(() => []),
  addToCart: jest.fn(),
  updateCartQuantity: jest.fn(),
  clearCart: jest.fn(),
}))

const api = require('../api/index.js')
const productApi = require('../api/product.js')

global.wx = {
  getStorageSync: jest.fn(() => ''),
  setStorageSync: jest.fn(),
  removeStorageSync: jest.fn(),
  reLaunch: jest.fn(),
  showToast: jest.fn(),
}
let pageConfig
global.Page = (config) => { pageConfig = config }
global.getApp = () => ({})

require('../pages/menu/index.js')

beforeEach(() => {
  jest.clearAllMocks()
})

describe('menu — 外卖冷启动（无 shop_id）', () => {
  it('onLoad order_type=delivery 且无 shop_id 时，在页内解析门店（不直接 loadData）', () => {
    const ctx = { setData: jest.fn(), loadDeliveryShop: jest.fn(), loadData: jest.fn() }
    pageConfig.onLoad.call(ctx, { order_type: 'delivery' })
    expect(ctx.loadDeliveryShop).toHaveBeenCalled()
    expect(ctx.loadData).not.toHaveBeenCalled()
  })

  it('loadDeliveryShop 成功：解析到门店并交给 loadData，不再二次 getShop', async () => {
    api.resolveDeliveryShop.mockResolvedValue({ id: 7, name: '鸡福旺' })
    const ctx = { data: {}, setData: jest.fn(), loadData: jest.fn() }
    await pageConfig.loadDeliveryShop.call(ctx)
    expect(api.resolveDeliveryShop).toHaveBeenCalled()
    expect(ctx.setData).toHaveBeenCalledWith(
      expect.objectContaining({ boundShopId: 7, boundTableNo: '', orderType: 'delivery' })
    )
    expect(ctx.loadData).toHaveBeenCalledWith({ id: 7, name: '鸡福旺' })
  })

  it('loadDeliveryShop 失败：error 态 + toast，不调用 loadData', async () => {
    api.resolveDeliveryShop.mockRejectedValue(new Error('no available shop'))
    const ctx = { data: {}, setData: jest.fn(), loadData: jest.fn() }
    await pageConfig.loadDeliveryShop.call(ctx)
    expect(ctx.setData).toHaveBeenCalledWith(
      expect.objectContaining({ loading: false, error: true })
    )
    expect(wx.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '暂无可配送门店' })
    )
    expect(ctx.loadData).not.toHaveBeenCalled()
  })

  it('loadData 收到预取门店时跳过 getShop，仅拉商品', async () => {
    productApi.getShopProducts.mockResolvedValue([])
    const ctx = { data: { boundShopId: 7 }, setData: jest.fn(), updateCartInfo: jest.fn() }
    await pageConfig.loadData.call(ctx, { id: 7, name: '鸡福旺' })
    expect(api.getShop).not.toHaveBeenCalled()
    expect(productApi.getShopProducts).toHaveBeenCalledWith(7)
    const shopSet = ctx.setData.mock.calls.find(([a]) => a && a.shop)
    expect(shopSet && shopSet[0].shop).toEqual({ id: 7, name: '鸡福旺' })
  })

  it('loadData 无预取门店时维持原行为：getShop 拉门店', async () => {
    api.getShop.mockResolvedValue({ id: 3, name: '堂食店' })
    productApi.getShopProducts.mockResolvedValue([])
    const ctx = { data: { boundShopId: 3 }, setData: jest.fn(), updateCartInfo: jest.fn() }
    await pageConfig.loadData.call(ctx)
    expect(api.getShop).toHaveBeenCalledWith(3)
    expect(productApi.getShopProducts).toHaveBeenCalledWith(3)
  })

  it('onRetry 在外卖未绑定门店时重走解析', () => {
    const ctx = { data: { orderType: 'delivery', boundShopId: 0 }, loadDeliveryShop: jest.fn(), loadData: jest.fn() }
    pageConfig.onRetry.call(ctx)
    expect(ctx.loadDeliveryShop).toHaveBeenCalled()
    expect(ctx.loadData).not.toHaveBeenCalled()
  })

  it('onRetry 在已绑定门店时正常重载', () => {
    const ctx = { data: { orderType: 'delivery', boundShopId: 7 }, loadDeliveryShop: jest.fn(), loadData: jest.fn() }
    pageConfig.onRetry.call(ctx)
    expect(ctx.loadData).toHaveBeenCalled()
    expect(ctx.loadDeliveryShop).not.toHaveBeenCalled()
  })
})
