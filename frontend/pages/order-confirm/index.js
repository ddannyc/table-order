// pages/order-confirm/index.js
const { getShop, getRewardBalance, createOrder, getDeliveryQuote, getTableBinding, setTableBinding } = require('../../api/index.js')
const { getCart, clearCart, getCartTotal } = require('../../api/product.js')
const { doLogin, handleAuthError, getLastDeliveryAddress, setLastDeliveryAddress } = require('../../utils/storage.js')

Page({
  data: {
    needLogin: false,
    shopId: 0,
    tableNo: '',
    orderType: 'dine_in',
    orderTypeLabel: '堂食',
    deliveryAddress: null,
    shop: {},
    cart: [],
    cartItems: [],
    rewardBalance: 0,
    useReward: false,
    loading: false,
    totalAmount: '0.00',
    actualPayAmount: '0.00',
    rewardDeduct: '0.00',
    rewardCeiling: '80',
    deliveryFee: '0.00',
    quoteToken: ''
  },

  onLoad(options) {
    const orderType = (options && options.order_type) || 'dine_in'
    this.setData({
      orderType,
      orderTypeLabel: orderType === 'delivery' ? '外卖' : '堂食',
      deliveryAddress: orderType === 'delivery' ? getLastDeliveryAddress() : null
    })
    if (options.shop_id) {
      this.setData({
        shopId: Number(options.shop_id) || 1,
        tableNo: options.table_no || ''
      })
      this.refreshCartDisplay()
      this.checkAuthAndLoad()
    } else {
      const { shopId, tableNo } = getTableBinding()
      if (shopId) {
        this.setData({ shopId, tableNo })
        this.refreshCartDisplay()
        this.checkAuthAndLoad()
      } else {
        wx.showModal({
          title: '请先扫码绑定桌号',
          content: '点击确定扫描餐桌二维码',
          success: (res) => {
            if (res.confirm) {
              this.scanQR()
            } else {
              wx.switchTab({ url: '/pages/home/index' })
            }
          }
        })
      }
    }
  },

  checkAuthAndLoad() {
    const token = wx.getStorageSync('token')
    if (!token) {
      this.setData({ needLogin: true })
      return
    }
    this.setData({ needLogin: false })
    this.loadData()
  },

  handleLogin() {
    doLogin()
      .then(() => {
        this.setData({ needLogin: false })
        this.loadData()
      })
      .catch(() => {})
  },

  onShow() {
    if (this.data.shopId) {
      this.refreshCartDisplay()
    }
  },

  refreshCartDisplay() {
    const cart = getCart(this.data.shopId, this.data.orderType)
    const total = getCartTotal(this.data.shopId, this.data.orderType)
    const cartItems = cart.map(item => ({
      ...item,
      subtotal: (Number(item.price) * Number(item.quantity)).toFixed(2)
    }))
    const rewardBalance = Number(this.data.rewardBalance) || 0
    const { useReward, shop, orderType } = this.data
    const rewardCeiling = Number(shop.reward_ceiling || 0.5) || 0.5
    let rewardDeduct = 0
    if (useReward && rewardBalance > 0) {
      rewardDeduct = Math.min(rewardBalance, total * rewardCeiling)
    }
    // Delivery fee (already quoted) is added on top of the item net; it never
    // enters the reward deduction base.
    const deliveryFee = orderType === 'delivery' ? (Number(this.data.deliveryFee) || 0) : 0
    const actualPay = total - rewardDeduct + deliveryFee
    const rewardCeilingPercent = (rewardCeiling * 100).toFixed(0)
    this.setData({
      cart, cartItems,
      totalAmount: total.toFixed(2),
      actualPayAmount: actualPay.toFixed(2),
      rewardDeduct: rewardDeduct.toFixed(2),
      rewardCeiling: rewardCeilingPercent
    })
  },

  loadData() {
    const { shopId } = this.data
    Promise.all([
      getShop(shopId),
      getRewardBalance()
    ]).then(([shop, rewardData]) => {
      const rb = Number(rewardData.reward_balance) || 0
      this.setData({
        shop,
        rewardBalance: rb
      })
      // Delivery with a cached address: refresh the quote now that we're authed.
      if (this.data.orderType === 'delivery' && this.data.deliveryAddress) {
        this.fetchQuote(this.data.deliveryAddress)
      } else {
        this.refreshCartDisplay()
      }
    }).catch(err => {
      if (!handleAuthError(err, this)) { console.error(err) }
    })
  },

  // Delivery: pick the WeChat address (name/phone/region text), then confirm the
  // exact delivery point on the map. chooseLocation returns the RECIPIENT's
  // coordinates — chooseAddress does not, and the device's own location (the old
  // wx.getLocation) is wrong whenever the buyer isn't standing at the address,
  // which made every far address quote against the phone's position.
  chooseDeliveryAddress() {
    wx.chooseAddress({
      success: (addr) => {
        wx.chooseLocation({
          success: (loc) => this.applyDeliveryAddress(addr, loc.latitude, loc.longitude),
          fail: (err) => {
            if (err && /cancel/.test(err.errMsg || '')) return // user backed out of the map
            wx.showModal({
              title: '需要选择收货位置',
              content: '请在地图上确认收货地点以计算配送费',
              showCancel: false
            })
          }
        })
      },
      fail: () => {}
    })
  },

  applyDeliveryAddress(addr, lat, lng) {
    const deliveryAddress = {
      userName: addr.userName,
      telNumber: addr.telNumber,
      provinceName: addr.provinceName,
      cityName: addr.cityName,
      countyName: addr.countyName,
      detailInfo: addr.detailInfo,
      lat,
      lng
    }
    setLastDeliveryAddress(deliveryAddress)
    this.setData({ deliveryAddress })
    this.fetchQuote(deliveryAddress)
  },

  fetchQuote(addr) {
    getDeliveryQuote(this.data.shopId, {
      recipient_name: addr.userName,
      recipient_phone: addr.telNumber,
      recipient_address: `${addr.provinceName || ''}${addr.cityName || ''}${addr.countyName || ''}${addr.detailInfo || ''}`,
      lat: addr.lat,
      lng: addr.lng
    }).then((res) => {
      this.setData({
        deliveryFee: (Number(res.delivery_fee) || 0).toFixed(2),
        quoteToken: res.quote_token || ''
      })
      this.refreshCartDisplay()
    }).catch((err) => {
      this.setData({ deliveryFee: '0.00', quoteToken: '' })
      if (!handleAuthError(err, this)) {
        wx.showToast({ title: (err && err.error) || '配送报价失败', icon: 'none' })
      }
      this.refreshCartDisplay()
    })
  },

  toggleUseReward() {
    const newUseReward = !this.data.useReward
    this.setData({ useReward: newUseReward })
    this.refreshCartDisplay()
  },

  handlePay() {
    if (this.data.loading) return
    const token = wx.getStorageSync('token')
    if (!token) {
      this.setData({ needLogin: true })
      return
    }
    const { cart, totalAmount, orderType, deliveryAddress, quoteToken } = this.data
    if (cart.length === 0) {
      wx.showToast({ title: '购物车为空', icon: 'none' })
      return
    }
    if (orderType === 'delivery' && (!deliveryAddress || !quoteToken)) {
      wx.showToast({ title: '请先选择收货地址', icon: 'none' })
      return
    }
    this.setData({ loading: true })
    const orderItems = cart.map(item => ({
      product_id: item.productId,
      spec_id: item.specId || 0,
      quantity: item.quantity
    }))
    let delivery = null
    if (orderType === 'delivery' && deliveryAddress) {
      delivery = {
        recipient_name: deliveryAddress.userName,
        recipient_phone: deliveryAddress.telNumber,
        province: deliveryAddress.provinceName,
        city: deliveryAddress.cityName,
        county: deliveryAddress.countyName,
        detail_address: deliveryAddress.detailInfo,
        lat: deliveryAddress.lat,
        lng: deliveryAddress.lng
      }
    }
    createOrder(this.data.shopId, this.data.tableNo, parseFloat(totalAmount), orderItems, this.data.useReward, orderType, delivery, quoteToken)
      .then((res) => {
        if (res.error) {
          wx.showToast({ title: res.error === 'prepay failed' ? '支付配置异常' : '下单失败', icon: 'none' })
          return
        }
        // Zero-amount order: skip WeChat Pay, go straight to success
        if (res.status === 2) {
          clearCart(this.data.shopId, this.data.orderType)
          wx.showModal({
            title: '下单成功',
            content: '福利金已全额抵扣，可在"我的"页面查看订单详情',
            showCancel: false,
            success: () => {
              wx.reLaunch({ url: '/pages/profile/index' })
            }
          })
          return
        }

        // Call WeChat Pay
        wx.requestPayment({
          timeStamp: res.time_stamp,
          nonceStr: res.nonce_str,
          package: res.package,
          signType: res.sign_type,
          paySign: res.pay_sign,
          success: () => {
            clearCart(this.data.shopId, this.data.orderType)
            wx.showModal({
              title: '支付成功',
              content: '订单已提交，可在"我的"页面查看订单详情',
              showCancel: false,
              success: () => {
                wx.reLaunch({ url: '/pages/profile/index' })
              }
            })
          },
          fail: (err) => {
            console.error(err)
            wx.showToast({ title: '支付取消或失败', icon: 'none' })
          }
        })
      })
      .catch((err) => {
        if (!handleAuthError(err, this)) {
          console.error(err)
          wx.showToast({ title: '下单失败', icon: 'none' })
        }
      })
      .finally(() => {
        this.setData({ loading: false })
      })
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
        this.setData({ shopId, tableNo })
        this.loadData()
      },
      fail: () => {
        wx.showToast({ title: '扫码失败', icon: 'none' })
      }
    })
  }
})