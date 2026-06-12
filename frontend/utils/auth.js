/**
 * Auth 工具函数
 * 统一的认证逻辑：登录、登出检测、401 恢复、return_path 守卫
 */

const { loginByCode, bindInviteCode } = require('../api/index.js')

// 直接使用 wx API，避免与 storage.js 的循环依赖

/**
 * 检查是否已登录（仅判断本地 token 是否存在，不验证有效期）
 */
export function isLoggedIn() {
  return !!wx.getStorageSync('token')
}

/**
 * 守卫：未登录时存储 return_path 并跳转到登录页。
 * 返回 true 表示已登录可继续，false 表示正在跳转。
 *
 * 用法：if (!requireLogin()) return;
 */
export function requireLogin() {
  if (isLoggedIn()) return true
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1]
  if (currentPage) {
    const route = '/' + currentPage.route
    const options = currentPage.options || {}
    const query = Object.keys(options)
      .map(k => k + '=' + options[k])
      .join('&')
    wx.setStorageSync('return_path', route + (query ? '?' + query : ''))
  }
  wx.navigateTo({ url: '/pages/login/index' })
  return false
}

/**
 * 触发微信登录 → 后端换取 token → 存储凭据 → 绑定待处理邀请码。
 * 返回 Promise，resolve 时登录已完成（token 已存储）。
 * 调用方负责刷新页面数据和 UI 状态。
 */
export function doLogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: (loginRes) => {
        loginByCode(loginRes.code)
          .then((data) => {
            if (!data.token) {
              wx.showToast({ title: '登录失败', icon: 'none' })
              reject(new Error('登录失败：token 缺失'))
              return
            }
            wx.setStorageSync('token', data.token)
            wx.setStorageSync('user', data.user || {})

            // 绑定待处理的邀请码（扫码未登录场景）
            const pendingCode = wx.getStorageSync('pending_invite_code')
            if (pendingCode) {
              wx.removeStorageSync('pending_invite_code')
              bindInviteCode(pendingCode)
                .catch(err => console.error('bind pending invite failed:', err))
                .finally(() => resolve(data))
            } else {
              resolve(data)
            }
          })
          .catch((err) => {
            wx.showToast({ title: '登录失败', icon: 'none' })
            reject(err)
          })
      },
      fail: () => {
        wx.showToast({ title: '微信登录失败', icon: 'none' })
        reject(new Error('wx.login failed'))
      }
    })
  })
}

/**
 * 共享 401 恢复：检测 API 错误是否为鉴权失败，并更新页面的 needLogin 状态。
 * 返回 true 表示已处理（是 401 且已设置 needLogin），调用方无需额外处理。
 *
 * 用法：apiCall().catch(err => { if (!handleAuthError(err, this)) { ... } })
 */
export function handleAuthError(err, page) {
  if (err && err.message === '未登录') {
    page.setData({ needLogin: true })
    return true
  }
  return false
}
