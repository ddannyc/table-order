// pages/home/index.js — 选餐入口启动页（堂食 / 外卖）
const { setTableBinding, bindInviteCode } = require('../../api/index.js')
const { TAB_LIST } = require('../../utils/tabbar.js')

Page({
  data: {
    tabbarList: TAB_LIST,
    tabbarCurrent: 0
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

  // 外卖：配送能力（闪送）未完备，暂作「即将上线」占位
  chooseDelivery() {
    wx.showModal({
      title: '外卖即将上线',
      content: '外卖配送功能正在开发中，敬请期待',
      showCancel: false
    })
  },

  tabChange(e) {
    const index = e.detail.index
    this.setData({ tabbarCurrent: index })
    const routes = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
    wx.reLaunch({ url: routes[index] + '?fromTabbar=1' })
  }
})
