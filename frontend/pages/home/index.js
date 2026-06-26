// pages/home/index.js — 选餐入口启动页（堂食 / 外卖；v6 设计稿）
const { setTableBinding, bindInviteCode, resolveDeliveryShop } = require('../../api/index.js')

Page({
  data: {
    tabbarCurrent: 0,
    mode: 'dine_in' // 分段视觉态：dine_in | delivery（实际动作直接触发，不依赖此态）
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
    }
  },

  // 分段「堂食」：仅切视觉态，实际开点走扫码卡
  selectDineIn() {
    this.setData({ mode: 'dine_in' })
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

  // 外卖：分段「外卖」即解析配送门店 → 进入菜单的外卖模式（无桌号）
  chooseDelivery() {
    this.setData({ mode: 'delivery' })
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
