// pages/home/index.js — 选餐入口启动页（堂食 / 外卖）
const { setTableBinding, bindInviteCode } = require('../../api/index.js')

Page({
  data: {
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

  // 外卖：Phase 4 接入，先占位
  chooseDelivery() {
    wx.showModal({
      title: '外卖即将上线',
      content: '外卖配送功能正在开发中，敬请期待',
      showCancel: false
    })
  },

  tabChange(e) {
    const index = e.detail.index
    this.setData({ 'tabbar.current': index })
    const routes = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
    wx.reLaunch({ url: routes[index] + '?fromTabbar=1' })
  }
})
