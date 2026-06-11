// pages/home/index.js
const { getShop, getTableBinding, setTableBinding, bindInviteCode } = require('../../api/index.js')
const { getShopProducts, getCartCount, getCartTotal, getCart, addToCart, updateCartQuantity } = require('../../api/product.js')

Page({
  data: {
    boundShopId: 0,
    boundTableNo: '',
    shop: {},
    products: [],
    categories: [],
    productsByCategory: {},
    cartCount: 0,
    cartTotal: '0.00',
    cartMap: {},
    cartQtyMap: {},
    loading: true,
    error: false,
    activeCategory: '',
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
    // Bind invite code from share card
    if (options && options.invite_code) {
      bindInviteCode(options.invite_code).catch(err => console.error('bind invite failed:', err))
    }
    this.checkTableBinding()
  },

  onShow() {
    if (this.data.boundShopId) {
      this.updateCartInfo()
    }
  },

  tabChange(e) {
    const index = e.detail.index
    this.setData({ 'tabbar.current': index })
    const routes = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
    wx.reLaunch({ url: routes[index] + '?fromTabbar=1' })
  },

  checkTableBinding() {
    const { shopId, tableNo } = getTableBinding()
    if (shopId && tableNo) {
      this.setData({ boundShopId: shopId, boundTableNo: tableNo })
      this.loadData()
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
      const activeCategory = categories.length > 0 ? categories[0] : ''
      this.setData({ shop, products, categories, productsByCategory, loading: false, activeCategory })
      this.updateCartInfo()
      this.setupCategoryObserver()
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

  getCartQty(productId) {
    return this.data.cartMap[productId] || 0
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

  changeTable() {
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

  onCategoryTap(e) {
    const cat = e.currentTarget.dataset.cat
    this.setData({ activeCategory: cat })
    const query = wx.createSelectorQuery()
    query.select('#panel-' + cat).boundingClientRect()
    query.selectViewport().scrollOffset()
    query.exec((res) => {
      if (res[0] && res[1]) {
        wx.pageScrollTo({ scrollTop: res[1].scrollTop + res[0].top - 80, duration: 300 })
      }
    })
  },

  setupCategoryObserver() {
    const observer = wx.createIntersectionObserver(this, { observeAll: true })
    this.data.categories.forEach(cat => {
      observer.relativeToViewport({ top: 80, bottom: 80 }).observe('#panel-' + cat, (res) => {
        if (res.intersectionRatio > 0) {
          this.setData({ activeCategory: cat })
        }
      })
    })
  },

  onRetry() {
    this.loadData()
  },

  onImgError(e) {
    const id = e.currentTarget.dataset.id
    const key = `productsByCategory`
    // mark image as failed so fallback shows
    const fallback = '/assets/img-fallback.png'
    e.target.src = fallback
  },

  goCart() {
    const { boundShopId, boundTableNo } = this.data
    wx.navigateTo({
      url: `/pages/order-confirm/index?shop_id=${boundShopId}&table_no=${boundTableNo}`
    })
  }
})