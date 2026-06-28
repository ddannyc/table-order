// pages/home/index.js — 选餐入口启动页（堂食 / 外卖；v6 设计稿）
const { setTableBinding, bindInviteCode } = require('../../api/index.js')

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

  // 外卖：点按即跳转菜单外卖模式（无桌号）；门店解析下沉到菜单页，
  // 由菜单的「加载中」态盖住等待，避免导航前的静默请求造成白屏卡顿。
  chooseDelivery() {
    wx.reLaunch({ url: '/pages/menu/index?order_type=delivery' })
  },

  tabChange(e) {
    const index = e.detail.index
    this.setData({ tabbarCurrent: index })
    const routes = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
    wx.reLaunch({ url: routes[index] + '?fromTabbar=1' })
  }
})
