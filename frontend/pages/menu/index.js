// pages/menu/index.js — 点餐菜单页（左分类栏 + 右列表）
const { getShop, getTableBinding, setTableBinding, clearTableBinding, bindInviteCode } = require('../../api/index.js')
const { getShopProducts, getCart, addToCart, updateCartQuantity, clearCart } = require('../../api/product.js')
const { resolveProductImage } = require('../../utils/menu-image.js')
const { pickDefaultSpec, clampQty, specPickerState } = require('../../utils/spec.js')
const { buildCartItems } = require('../../utils/cart-view.js')

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
    cartItems: [],         // 购物车弹层行（名/规/价/数量）
    cartSheetVisible: false,
    specPickerVisible: false,
    specPickerProduct: null,
    selectedSpecId: 0,    // 弹层当前选中规格
    specQty: 1,           // 弹层数量
    specUnitText: '0.00', // 选中规格单价
    specTotalText: '0.00',// 单价 × 数量
    specCanAdd: false,    // 选中在售才可加入
    loading: true,
    error: false,
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
    const patch = {
      cartQtyByKey, cartQtyByProduct, cartCount,
      cartTotal: cartTotal.toFixed(2),
      cartItems: buildCartItems(cart)
    }
    // 车空则关闭弹层（最后一件减没后不留空层）
    if (cartCount === 0) patch.cartSheetVisible = false
    this.setData(patch)
  },

  // 点击购物车区域 → 打开购物车弹层（车空不开）
  openCartSheet() {
    if (this.data.cartCount > 0) this.setData({ cartSheetVisible: true })
  },

  closeCartSheet() {
    this.setData({ cartSheetVisible: false })
  },

  // 弹层内：对某行 +1
  cartInc(e) {
    const key = e.currentTarget.dataset.key
    const qty = (this.data.cartQtyByKey[key] || 0) + 1
    updateCartQuantity(this.data.boundShopId, key, qty, this.data.orderType)
    this.updateCartInfo()
  },

  // 弹层内：对某行 -1（减到 0 删行；删空由 updateCartInfo 关层）
  cartDec(e) {
    const key = e.currentTarget.dataset.key
    const qty = (this.data.cartQtyByKey[key] || 0) - 1
    updateCartQuantity(this.data.boundShopId, key, qty, this.data.orderType)
    this.updateCartInfo()
  },

  // 清空当前购物车并关层
  clearCartAll() {
    clearCart(this.data.boundShopId, this.data.orderType)
    this.updateCartInfo()
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

  // 有规格商品：打开规格选择层（默认选首个在售规格，数量 1）
  openSpecPicker(e) {
    const productId = e.currentTarget.dataset.id
    const product = this.data.products.find(p => String(p.id) === String(productId))
    if (!product) return
    const specs = (product.specs || []).map(s => ({ ...s, priceText: s.price.toFixed(2) }))
    const def = pickDefaultSpec(specs)
    this.setData({
      specPickerVisible: true,
      specPickerProduct: { ...product, specs },
      selectedSpecId: def ? def.id : 0,
      specQty: 1
    })
    this.refreshSpecState()
  },

  closeSpecPicker() {
    this.setData({ specPickerVisible: false })
  },

  // 切换选中规格（售罄不可选）
  selectSpec(e) {
    const specId = e.currentTarget.dataset.specId
    const product = this.data.specPickerProduct
    if (!product) return
    const spec = product.specs.find(s => String(s.id) === String(specId))
    if (!spec || spec.status !== 1) return
    this.setData({ selectedSpecId: spec.id })
    this.refreshSpecState()
  },

  specDec() {
    this.setData({ specQty: clampQty(this.data.specQty - 1) })
    this.refreshSpecState()
  },

  specInc() {
    this.setData({ specQty: clampQty(this.data.specQty + 1) })
    this.refreshSpecState()
  },

  // 同步选中态：单价 / 合计 / 能否加入
  refreshSpecState() {
    const product = this.data.specPickerProduct
    const specs = product ? product.specs : []
    const st = specPickerState(specs, this.data.selectedSpecId, this.data.specQty)
    this.setData({
      selectedSpecId: st.selectedSpecId || 0,
      specUnitText: st.unitText,
      specTotalText: st.totalText,
      specCanAdd: st.canAdd
    })
  },

  // 按所选规格 × 数量加入购物车
  confirmAddSpec() {
    const product = this.data.specPickerProduct
    if (!product || !this.data.specCanAdd) return
    const spec = product.specs.find(s => String(s.id) === String(this.data.selectedSpecId))
    if (!spec || spec.status !== 1) return
    addToCart(this.data.boundShopId, product, spec, clampQty(this.data.specQty), this.data.orderType)
    this.updateCartInfo()
    wx.showToast({ title: '已加入', icon: 'success' })
    this.setData({ specPickerVisible: false })
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
