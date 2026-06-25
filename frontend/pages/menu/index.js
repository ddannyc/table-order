// pages/menu/index.js — 点餐菜单页（左分类栏 + 右列表）
const { getShop, getTableBinding, setTableBinding, bindInviteCode } = require('../../api/index.js')
const { getShopProducts, getCart, addToCart, updateCartQuantity } = require('../../api/product.js')

Page({
  data: {
    boundShopId: 0,
    boundTableNo: '',
    shop: {},
    products: [],
    categories: [],
    productsByCategory: {},
    activeCategory: '',
    scrollIntoId: '',
    orderType: 'dine_in', // dine_in | delivery（事实来源在首页/菜单，订单确认页只读）
    cartCount: 0,
    cartTotal: '0.00',
    cartMap: {},
    cartQtyMap: {},
    loading: true,
    error: false,
    tabbar: {
      current: 0,
      list: [
        { text: '点餐', key: 'home' },
        { text: '邀请', key: 'invite' },
        { text: '我的', key: 'profile' }
      ]
    }
  },

  onLoad(options) {
    if (options && options.invite_code) {
      bindInviteCode(options.invite_code).catch(err => console.error('bind invite failed:', err))
    }
    if (options && options.order_type) {
      this.setData({ orderType: options.order_type })
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
      categories.forEach(cat => {
        productsByCategory[cat] = products.filter(p => p.category === cat).map(p => ({
          ...p,
          priceText: p.price.toFixed(2)
        }))
      })
      this.setData({
        shop, products, categories, productsByCategory,
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
    const { boundShopId } = this.data
    const cart = getCart(boundShopId)
    let cartMap = {}
    let cartCount = 0
    let cartTotal = 0
    cart.forEach(item => {
      cartMap[item.id] = item.quantity
      cartCount += item.quantity
      cartTotal += item.price * item.quantity
    })
    this.setData({ cartMap, cartQtyMap: cartMap, cartCount, cartTotal: cartTotal.toFixed(2) })
  },

  // 左侧分类点击 → 高亮 + 右侧滚动到锚点
  selectCategory(e) {
    const { cat, index } = e.currentTarget.dataset
    this.setData({ activeCategory: cat, scrollIntoId: 'cat-' + index })
  },

  // 顶部堂食/外卖切换；外卖暂未上线，先占位不切换（Phase 4 接入）
  switchOrderType(e) {
    const type = e.currentTarget.dataset.type
    if (type === 'delivery') {
      wx.showModal({
        title: '外卖即将上线',
        content: '外卖配送功能正在开发中，敬请期待',
        showCancel: false
      })
      return
    }
    this.setData({ orderType: 'dine_in' })
  },

  onAdd(e) {
    const productId = e.currentTarget.dataset.id
    const product = this.data.products.find(p => String(p.id) === String(productId))
    if (!product) return
    addToCart(this.data.boundShopId, product, 1)
    this.updateCartInfo()
    wx.showToast({ title: '已加入', icon: 'success' })
  },

  onInc(e) {
    const productId = e.currentTarget.dataset.id
    const product = this.data.products.find(p => String(p.id) === String(productId))
    if (!product) return
    const qty = (this.data.cartQtyMap[product.id] || 0) + 1
    updateCartQuantity(this.data.boundShopId, product.id, qty)
    this.updateCartInfo()
  },

  onDec(e) {
    const productId = e.currentTarget.dataset.id
    const product = this.data.products.find(p => String(p.id) === String(productId))
    if (!product) return
    const qty = this.data.cartQtyMap[product.id] || 0
    if (qty <= 1) {
      updateCartQuantity(this.data.boundShopId, product.id, 0)
    } else {
      updateCartQuantity(this.data.boundShopId, product.id, qty - 1)
    }
    this.updateCartInfo()
  },

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

  onImgError(e) {
    e.target.src = '/assets/img-fallback.png'
  },

  goCart() {
    const { boundShopId, boundTableNo, orderType } = this.data
    wx.navigateTo({
      url: `/pages/order-confirm/index?shop_id=${boundShopId}&table_no=${boundTableNo}&order_type=${orderType}`
    })
  },

  tabChange(e) {
    const index = e.detail.index
    this.setData({ 'tabbar.current': index })
    const routes = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
    wx.reLaunch({ url: routes[index] + '?fromTabbar=1' })
  }
})
