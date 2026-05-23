App({
  onLaunch() {
    // 检查登录状态
    const token = wx.getStorageSync('token')
    if (!token) {
      // 未登录，跳转到登录页
      wx.reLaunch({ url: '/pages/login/index' })
    }
  }
})