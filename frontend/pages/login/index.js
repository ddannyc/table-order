// pages/login/index.js
const { loginByCode } = require('../../api/index.js')

Page({
  data: {
    loading: false
  },

  onLoad() {
    const token = wx.getStorageSync('token')
    if (token) {
      wx.reLaunch({ url: '/pages/home/index' })
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
              wx.reLaunch({ url: '/pages/home/index' })
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
  }
})