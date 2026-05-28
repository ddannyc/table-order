// pages/order-confirm/index.js
const { getShop, getBalance, createOrder, getTableBinding } = require('../../api/index.js')
const { getCart, clearCart, getCartTotal } = require('../../api/product.js')

Page({
  data: {
    shopId: 0,
    tableNo: '',
    shop: {},
    cart: [],
    cartItems: [],
    balance: 0,
    rewardBalance: 0,
    useReward: false,
    loading: false,
    totalAmount: '0.00',
    actualPayAmount: '0.00',
    rewardDeduct: '0.00',
    availableBalance: '0.00',
    rewardCeiling: '80'
  },

  onLoad(options) {
    if (options.shop_id) {
      this.setData({
        shopId: Number(options.shop_id) || 1,
        tableNo: options.table_no || ''
      })
      this.loadData()
    } else {
      const { shopId, tableNo } = getTableBinding()
      if (shopId) {
        this.setData({ shopId, tableNo })
        this.loadData()
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

  onShow() {
    if (this.data.shopId) {
      this.refreshCartDisplay()
    }
  },

  refreshCartDisplay() {
    const cart = getCart(this.data.shopId)
    const total = getCartTotal(this.data.shopId)
    const cartItems = cart.map(item => ({
      ...item,
      subtotal: (Number(item.price) * Number(item.quantity)).toFixed(2)
    }))
    const balance = Number(this.data.balance) || 0
    const rewardBalance = Number(this.data.rewardBalance) || 0
    const { useReward, shop } = this.data
    const rewardCeiling = Number(shop.reward_ceiling || 0.5) || 0.5
    let rewardDeduct = 0
    let actualPay = total
    if (useReward && rewardBalance > 0) {
      rewardDeduct = Math.min(rewardBalance, total * rewardCeiling)
      actualPay = total - rewardDeduct
    }
    const availableBalance = balance + (useReward ? rewardBalance : 0)
    const rewardCeilingPercent = (rewardCeiling * 100).toFixed(0)
    this.setData({
      cart, cartItems,
      totalAmount: total.toFixed(2),
      actualPayAmount: actualPay.toFixed(2),
      rewardDeduct: rewardDeduct.toFixed(2),
      availableBalance: availableBalance.toFixed(2),
      rewardCeiling: rewardCeilingPercent
    })
  },

  loadData() {
    const { shopId } = this.data
    Promise.all([
      getShop(shopId),
      getBalance()
    ]).then(([shop, balanceData]) => {
      const b = Number(balanceData.balance) || 0
      const rb = Number(balanceData.reward_balance) || 0
      this.setData({
        shop,
        balance: b,
        rewardBalance: rb,
        availableBalance: b.toFixed(2)
      })
      this.refreshCartDisplay()
    }).catch(err => {
      console.error(err)
    })
  },

  toggleUseReward() {
    const newUseReward = !this.data.useReward
    this.setData({ useReward: newUseReward })
    this.refreshCartDisplay()
  },

  handlePay() {
    if (this.data.loading) return
    const { cart, totalAmount } = this.data
    if (cart.length === 0) {
      wx.showToast({ title: '购物车为空', icon: 'none' })
      return
    }
    const actualPay = parseFloat(this.data.actualPayAmount)
    const totalWithReward = Number(this.data.balance) + (this.data.useReward ? Number(this.data.rewardBalance) : 0)
    if (actualPay > totalWithReward) {
      wx.showToast({ title: '余额不足', icon: 'none' })
      return
    }
    this.setData({ loading: true })
    const orderItems = cart.map(item => ({
      product_id: item.id,
      quantity: item.quantity,
      price: Number(item.price) || 0
    }))
    createOrder(this.data.shopId, this.data.tableNo, parseFloat(totalAmount), orderItems, this.data.useReward)
      .then((res) => {
        if (res.id) {
          clearCart(this.data.shopId)
          wx.showModal({
            title: '支付成功',
            content: '订单已提交，可在"我的"页面查看订单详情',
            showCancel: false,
            success: () => {
              wx.reLaunch({ url: '/pages/profile/index' })
            }
          })
        }
      })
      .catch((err) => {
        console.error(err)
        wx.showToast({ title: '支付失败', icon: 'none' })
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
        this.setData({ shopId, tableNo })
        this.loadData()
      },
      fail: () => {
        wx.showToast({ title: '扫码失败', icon: 'none' })
      }
    })
  }
})