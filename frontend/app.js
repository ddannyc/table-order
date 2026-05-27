App({
  onLaunch(options) {
    // Handle scene from QR code scan (wxacode.getUnlimited)
    this.handleScene(options)
    // 检查登录状态
    const token = wx.getStorageSync('token')
    if (!token) {
      // 未登录，跳转到登录页
      wx.reLaunch({ url: '/pages/login/index' })
    }
  },

  onShow(options) {
    this.handleScene(options)
  },

  handleScene(options) {
    if (!options || !options.query) return
    const scene = options.query.scene
    if (!scene) return
    // scene format: ic=INVITE_CODE
    const match = scene.match(/^ic=(.+)$/)
    if (match) {
      const inviteCode = match[1]
      wx.setStorageSync('pending_invite_code', inviteCode)
      // Navigate to invite page to bind
      wx.reLaunch({ url: '/pages/invite/index' })
    }
  }
})