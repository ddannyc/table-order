// pages/invite/index.js
const { getInviteStats, generateInviteCode } = require('../../api/index.js')

Page({
  data: {
    stats: { invite_count: 0, total_invite_reward: 0, today_reward: 0 },
    statsText: { totalInviteReward: '0.00', todayReward: '0.00' },
    inviteURL: '',
    tabbar: {
      current: 1,
      list: [
        { text: '点餐', key: 'home' },
        { text: '邀请', key: 'invite' },
        { text: '我的', key: 'profile' }
      ]
    }
  },

  onLoad() {
    this.loadData()
  },

  onShareAppMessage() {
    // Ensure inviteURL exists before sharing
    if (!this.data.inviteURL) {
      generateInviteCode().then(res => {
        this.setData({ inviteURL: res.invite_url || '' })
      })
    }
    return {
      title: '来XX餐饮消费可返10%福利金',
      path: this.data.inviteURL || '/pages/home/index'
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
        if (url) wx.setStorageSync('invite_url', url)
        this.setData({ inviteURL: url })
      }).catch(err => {
        console.error(err)
      })
    }
  },

  copyLink() {
    if (!this.data.inviteURL) {
      wx.showToast({ title: '链接未生成', icon: 'none' })
      return
    }
    wx.setClipboardData({
      data: this.data.inviteURL,
      success: () => wx.showToast({ title: '已复制', icon: 'success' })
    })
  }
})