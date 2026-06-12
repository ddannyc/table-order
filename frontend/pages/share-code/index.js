// pages/share-code/index.js
const { getInviteQR, getRewardBalance } = require('../../api/index.js')
const { doLogin, handleAuthError } = require('../../utils/storage.js')

Page({
  data: {
    needLogin: false,
    qrCodeSrc: '',
    inviteURL: '',
    rewardPaused: false,
    rewardBalance: '0.00'
  },

  onLoad() {
    const token = wx.getStorageSync('token')
    if (!token) {
      this.setData({ needLogin: true })
      return
    }
    this.loadData()
  },

  handleLogin() {
    doLogin()
      .then(() => {
        this.setData({ needLogin: false })
        this.loadData()
      })
      .catch(() => {})
  },

  loadData() {
    // Load invite QR code from API (no local cache — always fetch fresh)
    getInviteQR().then(data => {
      const base64 = wx.arrayBufferToBase64(data)
      this.setData({ qrCodeSrc: 'data:image/jpeg;base64,' + base64 })
    }).catch(err => {
      if (!handleAuthError(err, this)) { console.error('load invite qr failed:', err) }
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
    }).catch(err => {
      if (!handleAuthError(err, this)) { console.error(err) }
    })
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
