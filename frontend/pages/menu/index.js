// pages/menu/index.js — 点餐菜单页（左分类栏 + 右列表）
const { getShop, getTableBinding, setTableBinding, clearTableBinding, bindInviteCode } = require('../../api/index.js')
const { getShopProducts, getCart, addToCart, updateCartQuantity } = require('../../api/product.js')
const { TAB_LIST } = require('../../utils/tabbar.js')
const { resolveProductImage } = require('../../utils/menu-image.js')

Page({
  data: {
    boundShopId: 0,
    boundTableNo: '',
    shop: {},
    products: [],
    categories: [],
    productsByCategory: {},
    categoryCounts: {},
    activeCategory: '',
    orderType: 'dine_in', // dine_in | delivery（事实来源在首页/菜单，订单确认页只读）
    cartCount: 0,
    cartTotal: '0.00',
    cartQtyByKey: {},      // `${productId}_${specId}` -> qty
    cartQtyByProduct: {},  // productId -> total qty (for spec products)
    specPickerVisible: false,
    specPickerProduct: null,
    loading: true,
    error: false,
    tabbarList: TAB_LIST,
    tabbarCurrent: 0
  },

  onLoad(options) {
    if (options && options.invite_code) {
      bindInviteCode(options.invite_code).catch(err => console.error('bind invite failed:', err))
    }
    if (options && options.order_type) {
      this.setData({ orderType: options.order_type })
    }
    // 外卖：带 shop_id、无桌号。清除历史堂食绑定，避免被误判回堂食。
    if (options && options.order_type === 'delivery' && options.shop_id) {
      clearTableBinding()
      this.setData({ boundShopId: Number(options.shop_id), boundTableNo: '', orderType: 'delivery' })
      this.loadData()
      return
    }
    if (options && options.shop_id && options.table_no) {
      const shopId = Number(options.shop_id)
      const tableNo = options.table_no
      setTableBinding(shopId, tableNo)
      this.setData({ boundShopId: shopId, boundTableNo: tableNo })
      this.loadData()
      return
    }
    this.checkTableBinding()
  },

  onShow() {
    // 配送态不跟随桌号绑定（避免被历史堂食绑定覆盖）
    if (this.data.orderType === 'delivery') {
      if (this.data.boundShopId) this.updateCartInfo()
      return
    }
    const { shopId, tableNo } = getTableBinding()
    if (shopId && tableNo && (shopId !== this.data.boundShopId || tableNo !== this.data.boundTableNo)) {
      this.setData({ boundShopId: shopId, boundTableNo: tableNo })
      this.loadData()
      return
    }
    if (this.data.boundShopId) {
      this.updateCartInfo()
    }
  },

  checkTableBinding() {
    const { shopId, tableNo } = getTableBinding()
    if (shopId && tableNo) {
      this.setData({ boundShopId: shopId, boundTableNo: tableNo })
      this.loadData()
    } else {
      this.setData({ loading: false })
    }
  },

  loadData() {
    const { boundShopId } = this.data
    if (!boundShopId) return
    this.setData({ loading: true, error: false })
    Promise.all([
      getShop(boundShopId),
      getShopProducts(boundShopId)
    ]).then(([shop, products]) => {
      const categories = [...new Set(products.map(p => p.category))]
      const productsByCategory = {}
      const categoryCounts = {}
      categories.forEach(cat => {
        productsByCategory[cat] = products.filter(p => p.category === cat).map(p => {
          const specs = (p.specs || []).map(s => ({ ...s, priceText: s.price.toFixed(2) }))
          const hasSpecs = specs.length > 0
          const img = resolveProductImage(p)
          return {
            ...p,
            specs,
            hasSpecs,
            noSpecKey: `${p.id}_0`,
            specMinText: hasSpecs ? Math.min(...specs.map(s => s.price)).toFixed(2) : null,
            priceText: p.price.toFixed(2),
            hasImage: img.hasImage,
            ph: img.placeholder
          }
        })
        categoryCounts[cat] = productsByCategory[cat].length
      })
      this.setData({
        shop, products, categories, productsByCategory, categoryCounts,
        activeCategory: categories[0] || '',
        loading: false
      })
      this.updateCartInfo()
    }).catch(err => {
      console.error(err)
      this.setData({ loading: false, error: true })
      wx.showToast({ title: '加载失败', icon: 'none' })
    })
  },

  updateCartInfo() {
    const { boundShopId, orderType } = this.data
    const cart = getCart(boundShopId, orderType)
    let cartQtyByKey = {}
    let cartQtyByProduct = {}
    let cartCount = 0
    let cartTotal = 0
    cart.forEach(item => {
      cartQtyByKey[item.key] = item.quantity
      cartQtyByProduct[item.productId] = (cartQtyByProduct[item.productId] || 0) + item.quantity
      cartCount += item.quantity
      cartTotal += item.price * item.quantity
    })
    this.setData({ cartQtyByKey, cartQtyByProduct, cartCount, cartTotal: cartTotal.toFixed(2) })
  },

  // 分类页签切换 → 显示该分类列表
  selectCategory(e) {
    this.setData({ activeCategory: e.currentTarget.dataset.cat })
  },

  // 无规格商品：直接加购（specId 0）
  onAdd(e) {
    const productId = e.currentTarget.dataset.id
    const product = this.data.products.find(p => String(p.id) === String(productId))
    if (!product) return
    addToCart(this.data.boundShopId, product, null, 1, this.data.orderType)
    this.updateCartInfo()
    wx.showToast({ title: '已加入', icon: 'success' })
  },

  onInc(e) {
    const productId = e.currentTarget.dataset.id
    const key = `${productId}_0`
    const qty = (this.data.cartQtyByKey[key] || 0) + 1
    updateCartQuantity(this.data.boundShopId, key, qty, this.data.orderType)
    this.updateCartInfo()
  },

  onDec(e) {
    const productId = e.currentTarget.dataset.id
    const key = `${productId}_0`
    const qty = this.data.cartQtyByKey[key] || 0
    updateCartQuantity(this.data.boundShopId, key, qty <= 1 ? 0 : qty - 1, this.data.orderType)
    this.updateCartInfo()
  },

  // 有规格商品：打开规格选择层
  openSpecPicker(e) {
    const productId = e.currentTarget.dataset.id
    const product = this.data.products.find(p => String(p.id) === String(productId))
    if (!product) return
    const specs = (product.specs || []).map(s => ({ ...s, priceText: s.price.toFixed(2) }))
    this.setData({ specPickerVisible: true, specPickerProduct: { ...product, specs } })
  },

  closeSpecPicker() {
    this.setData({ specPickerVisible: false })
  },

  pickSpec(e) {
    const specId = e.currentTarget.dataset.specId
    const product = this.data.specPickerProduct
    if (!product) return
    const spec = product.specs.find(s => String(s.id) === String(specId))
    if (!spec) return
    addToCart(this.data.boundShopId, product, spec, 1, this.data.orderType)
    this.updateCartInfo()
    wx.showToast({ title: '已加入', icon: 'success' })
  },

  noop() {},

  scanQR() {
    wx.scanCode({
      success: (res) => {
        const query = res.result.split('?')[1] || ''
        const params = {}
        query.split('&').forEach(pair => {
          const [key, val] = pair.split('=')
          if (key) params[key] = val || ''
        })
        const shopId = Number(params.shop_id) || 1
        const tableNo = params.table_no || 'A01'
        setTableBinding(shopId, tableNo)
        this.setData({ boundShopId: shopId, boundTableNo: tableNo })
        this.loadData()
      },
      fail: () => {
        wx.showToast({ title: '扫码失败', icon: 'none' })
      }
    })
  },

  goHome() {
    wx.reLaunch({ url: '/pages/home/index' })
  },

  onRetry() {
    this.loadData()
  },

  // 图片加载失败 → 回退到该分类的 CSS 占位块（不显示裂图）。
  onImgError(e) {
    const id = e.currentTarget.dataset.id
    const { productsByCategory, activeCategory } = this.data
    const updated = { ...productsByCategory }
    updated[activeCategory] = (updated[activeCategory] || []).map(p =>
      String(p.id) === String(id) ? { ...p, hasImage: false } : p
    )
    this.setData({ productsByCategory: updated })
  },

  goCart() {
    const { boundShopId, boundTableNo, orderType } = this.data
    wx.navigateTo({
      url: `/pages/order-confirm/index?shop_id=${boundShopId}&table_no=${boundTableNo}&order_type=${orderType}`
    })
  },

  tabChange(e) {
    const index = e.detail.index
    this.setData({ tabbarCurrent: index })
    const routes = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
    wx.reLaunch({ url: routes[index] + '?fromTabbar=1' })
  }
})
