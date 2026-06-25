App({
  onLaunch(options) {
    // Handle scene from QR code scan (wxacode.getUnlimited)
    this.handleScene(options)
  },

  onShow(options) {
    this.handleScene(options)
  },

  handleScene(options) {
    if (!options || !options.query) return

    // Handle URL Scheme query params (from /scan H5 redirect)
    // When mini-program opens via URL Scheme, params arrive in options.query
    const shopId = options.query.shop_id
    const tableNo = options.query.table_no
    if (shopId && tableNo) {
      wx.setStorageSync('current_shop_id', Number(shopId))
      wx.setStorageSync('current_table_no', tableNo)
      wx.removeStorageSync('pending_invite_code')
      wx.reLaunch({ url: `/pages/menu/index?shop_id=${shopId}&table_no=${tableNo}` })
      return
    }

    // Handle scene string from QR code scan.
    //   scene: from wxacode.getUnlimited QR, e.g. "shop_id=1&table_no=A01&token=xxx"
    //   q:     full URL from "扫普通链接二维码打开小程序" rule (URL-encoded),
    //          e.g. "https://host/scan?shop_id=1&table_no=A01&token=xxx" — no /scan
    //          request hits the backend in this path. The shop_id/table_no regexes
    //          below match inside the decoded URL just as they do a bare scene.
    const scene = options.query.scene || (options.query.q && decodeURIComponent(options.query.q))
    if (!scene) return

    // shop_id=XXX&table_no=YYY (table QR code via WeChat link rule)
    const shopMatch = scene.match(/shop_id=(\d+)/)
    const tableMatch = scene.match(/table_no=([^&\s]+)/)
    if (shopMatch && tableMatch) {
      const shopId = Number(shopMatch[1])
      const tableNo = tableMatch[1]
      wx.setStorageSync('current_shop_id', shopId)
      wx.setStorageSync('current_table_no', tableNo)
      wx.removeStorageSync('pending_invite_code')
      wx.reLaunch({ url: `/pages/menu/index?shop_id=${shopId}&table_no=${tableNo}` })
      return
    }

    // ic=INVITE_CODE (invite QR code via wxacode.getUnlimited)
    const codeMatch = scene.match(/^ic=(.+)$/)
    if (codeMatch) {
      const inviteCode = codeMatch[1]
      wx.setStorageSync('pending_invite_code', inviteCode)
      // Navigate to invite page to bind
      wx.reLaunch({ url: '/pages/invite/index' })
    }
  }
})