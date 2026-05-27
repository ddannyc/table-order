// pages/invite/index.js
const { getInviteStats, generateInviteCode, bindInviteCode, getInviteQR } = require('../../api/index.js')

Page({
  data: {
    stats: { invite_count: 0, total_invite_reward: 0, today_reward: 0 },
    statsText: { totalInviteReward: '0.00', todayReward: '0.00' },
    inviteURL: '',
    qrCodeSrc: '',
    tabbar: {
      current: 1,
      list: [
        { text: '点餐', key: 'home' },
        { text: '邀请', key: 'invite' },
        { text: '我的', key: 'profile' }
      ]
    }
  },

  onLoad(options) {
    // Check if opened from share link with invite_code
    if (options && options.invite_code) {
      bindInviteCode(options.invite_code).catch(err => console.error(err))
    }
    // Check if opened from QR code scan (pending invite from app.js handleScene)
    const pendingCode = wx.getStorageSync('pending_invite_code')
    if (pendingCode) {
      wx.removeStorageSync('pending_invite_code')
      bindInviteCode(pendingCode).catch(err => console.error(err))
    }
    this.loadData()
  },

  onShareAppMessage() {
    const cached = wx.getStorageSync('invite_url')
    let path = cached || '/pages/home/index'
    // Guard: cached URL may be old external URL from previous app version.
    // WeChat mini-program share path must be a page path starting with /.
    if (path.indexOf('http') === 0) {
      path = '/pages/home/index'
    }
    return {
      title: '来XX餐饮消费可返10%福利金',
      path
    }
  },

  tabChange(e) {
    const index = e.detail.index
    this.setData({ 'tabbar.current': index })
    const routes = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
    wx.reLaunch({ url: routes[index] + '?fromTabbar=1' })
  },

  loadData() {
    getInviteStats().then(stats => {
      this.setData({
        stats,
        statsText: {
          totalInviteReward: (stats.total_invite_reward || 0).toFixed(2),
          todayReward: (stats.today_reward || 0).toFixed(2)
        }
      })
    }).catch(err => {
      console.error(err)
    })
    // Generate invite URL only once and cache in storage
    const cached = wx.getStorageSync('invite_url')
    if (cached) {
      this.setData({ inviteURL: cached })
    } else {
      generateInviteCode().then(res => {
        const url = res.invite_url || ''
        if (url && url.indexOf('http') !== 0) {
          wx.setStorageSync('invite_url', url)
          this.setData({ inviteURL: url })
        } else if (url && url.indexOf('http') === 0) {
          // Backend returned external URL — regenerate to get correct page path
          wx.removeStorageSync('invite_url')
          this.loadData()
        }
      }).catch(err => {
        console.error(err)
      })
    }

    // Load invite QR code
    getInviteQR().then(data => {
      const base64 = wx.arrayBufferToBase64(data)
      this.setData({ qrCodeSrc: 'data:image/jpeg;base64,' + base64 })
    }).catch(err => {
      console.error('load invite qr failed:', err)
    })
  }
})