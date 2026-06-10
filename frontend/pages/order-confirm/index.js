// pages/order-confirm/index.js
const { getShop, getRewardBalance, createOrder, getTableBinding } = require('../../api/index.js')
const { getCart, clearCart, getCartTotal } = require('../../api/product.js')

Page({
  data: {
    shopId: 0,
    tableNo: '',
    shop: {},
    cart: [],
    cartItems: [],
    rewardBalance: 0,
    useReward: false,
    loading: false,
    totalAmount: '0.00',
    actualPayAmount: '0.00',
    rewardDeduct: '0.00',
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
    const rewardBalance = Number(this.data.rewardBalance) || 0
    const { useReward, shop } = this.data
    const rewardCeiling = Number(shop.reward_ceiling || 0.5) || 0.5
    let rewardDeduct = 0
    let actualPay = total
    if (useReward && rewardBalance > 0) {
      rewardDeduct = Math.min(rewardBalance, total * rewardCeiling)
      actualPay = total - rewardDeduct
    }
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
    this.setData({ loading: true })
    const orderItems = cart.map(item => ({
      product_id: item.id,
      quantity: item.quantity,
      price: Number(item.price) || 0
    }))
    createOrder(this.data.shopId, this.data.tableNo, parseFloat(totalAmount), orderItems, this.data.useReward)
      .then((res) => {
        if (res.error) {
          wx.showToast({ title: res.error === 'prepay failed' ? '支付配置异常' : '下单失败', icon: 'none' })
          return
        }
        // Zero-amount order: skip WeChat Pay, go straight to success
        if (res.status === 2) {
          clearCart(this.data.shopId)
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
            clearCart(this.data.shopId)
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
        console.error(err)
        wx.showToast({ title: '下单失败', icon: 'none' })
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