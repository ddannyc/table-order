// pages/home/index.js — 选餐入口启动页（堂食 / 外卖）
const { setTableBinding, bindInviteCode, resolveDeliveryShop, getBalance, getRewardBalance } = require('../../api/index.js')
const { TAB_LIST } = require('../../utils/tabbar.js')

Page({
  data: {
    tabbarList: TAB_LIST,
    tabbarCurrent: 0,
    balanceText: '—', // 余额（真数据；未登录显 —，不假数字）
    rewardText: '—'   // 返利余额
  },

  onLoad(options) {
    // 分享卡片带邀请码进入
    if (options && options.invite_code) {
      bindInviteCode(options.invite_code).catch(err => console.error('bind invite failed:', err))
    }
    // 深链带桌号（防御性：通常 app.js handleScene 已直达 menu）
    if (options && options.shop_id && options.table_no) {
      setTableBinding(Number(options.shop_id), options.table_no)
      wx.reLaunch({ url: '/pages/menu/index' })
      return
    }
    this.loadWallet()
  },

  onShow() {
    this.loadWallet()
  },

  // 品牌头的余额/返利：登录态走真接口；未登录显 —，绝不假数据
  loadWallet() {
    const token = wx.getStorageSync('token')
    if (!token) {
      this.setData({ balanceText: '—', rewardText: '—' })
      return Promise.resolve()
    }
    return Promise.all([getBalance(), getRewardBalance()])
      .then(([w, r]) => {
        this.setData({
          balanceText: Number((w && w.balance) || 0).toFixed(2),
          rewardText: Number((r && r.reward_balance) || 0).toFixed(2)
        })
      })
      .catch(() => this.setData({ balanceText: '—', rewardText: '—' }))
  },

  // 堂食：扫码绑桌 → 菜单
  scanDineIn() {
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
        wx.reLaunch({ url: '/pages/menu/index' })
      },
      fail: () => {
        wx.showToast({ title: '扫码失败', icon: 'none' })
      }
    })
  },

  // 外卖：解析配送门店 → 进入菜单的外卖模式（无桌号）
  chooseDelivery() {
    return resolveDeliveryShop()
      .then(shop => {
        wx.reLaunch({ url: `/pages/menu/index?order_type=delivery&shop_id=${shop.id}` })
      })
      .catch(() => {
        wx.showToast({ title: '暂无可配送门店', icon: 'none' })
      })
  },

  tabChange(e) {
    const index = e.detail.index
    this.setData({ tabbarCurrent: index })
    const routes = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
    wx.reLaunch({ url: routes[index] + '?fromTabbar=1' })
  }
})
