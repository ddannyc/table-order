// pages/share-code/index.js
const { getInviteQR, getRewardBalance } = require('../../api/index.js')

Page({
  data: {
    qrCodeSrc: '',
    inviteURL: '',
    rewardPaused: false,
    rewardBalance: '0.00'
  },

  onLoad() {
    this.loadData()
  },

  loadData() {
    // Load invite QR code from API (no local cache — always fetch fresh)
    getInviteQR().then(data => {
      const base64 = wx.arrayBufferToBase64(data)
      this.setData({ qrCodeSrc: 'data:image/jpeg;base64,' + base64 })
    }).catch(err => {
      console.error('load invite qr failed:', err)
    })

    const cached = wx.getStorageSync('invite_url')
    if (cached) {
      this.setData({ inviteURL: cached })
    } else {
      const user = wx.getStorageSync('user')
      if (user && user.invite_code) {
        const url = '/pages/home/index?invite_code=' + user.invite_code
        wx.setStorageSync('invite_url', url)
        this.setData({ inviteURL: url })
      }
    }

    getRewardBalance().then(rb => {
      this.setData({
        rewardPaused: rb.reward_paused || false,
        rewardBalance: (rb.reward_balance || 0).toFixed(2)
      })
    }).catch(() => {})
  },

  onShareAppMessage() {
    const cached = wx.getStorageSync('invite_url')
    let path = cached || '/pages/home/index'
    if (path.indexOf('http') === 0) {
      path = '/pages/home/index'
    }
    return {
      title: '来店消费返福利金，自购3%直推10%间推4%',
      path
    }
  },

  copyLink() {
    const user = wx.getStorageSync('user')
    if (user && user.invite_code) {
      wx.setClipboardData({
        data: user.invite_code,
        success: () => wx.showToast({ title: '已复制邀请码', icon: 'success' })
      })
    }
  }
})
