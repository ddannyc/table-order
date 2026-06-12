// pages/login/index.js
const { bindInviteCode } = require('../../api/index.js')
const { doLogin } = require('../../utils/storage.js')

Page({
  data: {
    loading: false
  },

  onLoad() {
    const token = wx.getStorageSync('token')
    if (token) {
      const returnPath = wx.getStorageSync('return_path')
      if (returnPath) {
        wx.removeStorageSync('return_path')
        wx.redirectTo({ url: returnPath })
      } else {
        this.redirectAfterLogin()
      }
    }
  },

  handleLogin() {
    if (this.data.loading) return
    this.setData({ loading: true })
    doLogin()
      .then(() => {
        this.setData({ loading: false })
        this.redirectAfterLogin()
      })
      .catch(() => {
        this.setData({ loading: false })
      })
  },

  /**
   * 绑定待处理邀请码。
   * doLogin() 内部已在登录时绑定 pending_invite_code，此处仅处理
   * "已有 token 直接进入登录页"的边界情况（无需重新登录）。
   */
  bindPendingInvite() {
    const pendingCode = wx.getStorageSync('pending_invite_code')
    if (pendingCode) {
      wx.removeStorageSync('pending_invite_code')
      bindInviteCode(pendingCode).catch(err => console.error('bind invite failed:', err))
    }
  },

  redirectAfterLogin() {
    // 优先返回来源页面
    const returnPath = wx.getStorageSync('return_path')
    if (returnPath) {
      wx.removeStorageSync('return_path')
      this.bindPendingInvite()
      wx.redirectTo({ url: returnPath })
      return
    }
    // 有待绑定邀请码 → 去邀请页
    const pendingCode = wx.getStorageSync('pending_invite_code')
    if (pendingCode) {
      this.bindPendingInvite()
      wx.reLaunch({ url: '/pages/invite/index' })
    } else {
      wx.reLaunch({ url: '/pages/home/index' })
    }
  }
})
