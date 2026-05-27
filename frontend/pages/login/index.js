// pages/login/index.js
const { loginByCode, bindInviteCode } = require('../../api/index.js')

Page({
  data: {
    loading: false
  },

  onLoad() {
    const token = wx.getStorageSync('token')
    if (token) {
      this.redirectAfterLogin()
    }
  },

  handleLogin() {
    if (this.data.loading) return
    this.setData({ loading: true })

    wx.login({
      success: (loginRes) => {
        loginByCode(loginRes.code)
          .then((res) => {
            if (res.token) {
              wx.setStorageSync('token', res.token)
              wx.setStorageSync('user', res.user || {})
              this.bindPendingInvite()
              this.redirectAfterLogin()
            } else {
              wx.showToast({ title: '登录失败', icon: 'none' })
            }
          })
          .catch((err) => {
            console.error(err)
            wx.showToast({ title: '登录失败', icon: 'none' })
          })
          .finally(() => {
            this.setData({ loading: false })
          })
      },
      fail: () => {
        wx.showToast({ title: '微信登录失败', icon: 'none' })
        this.setData({ loading: false })
      }
    })
  },

  bindPendingInvite() {
    const pendingCode = wx.getStorageSync('pending_invite_code')
    if (pendingCode) {
      wx.removeStorageSync('pending_invite_code')
      bindInviteCode(pendingCode).catch(err => console.error('bind invite failed:', err))
    }
  },

  redirectAfterLogin() {
    const pendingCode = wx.getStorageSync('pending_invite_code')
    if (pendingCode) {
      this.bindPendingInvite()
      wx.reLaunch({ url: '/pages/invite/index' })
    } else {
      wx.reLaunch({ url: '/pages/home/index' })
    }
  }
})