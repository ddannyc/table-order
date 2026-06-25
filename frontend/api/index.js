/**
 * API 客户端层
 * 替换 uni.request → wx.request，函数签名保持不变
 */

const { API_BASE } = require('../config.js')

export const request = (options) => {
  return new Promise((resolve, reject) => {
    const token = (options.auth !== false) ? (wx.getStorageSync('token') || '') : ''
    const url = API_BASE + options.url
    wx.request({
      url,
      method: options.method || 'GET',
      data: options.data || {},
      responseType: options.responseType || undefined,
      header: Object.assign(
        { 'Content-Type': 'application/json' },
        token ? { 'Authorization': `Bearer ${token}` } : {}
      ),
      success: (res) => {
        if (res.statusCode === 200) {
          resolve(res.data)
        } else if (res.statusCode === 401) {
          wx.removeStorageSync('token')
          reject(new Error('未登录'))
        } else {
          reject(res.data)
        }
      },
      fail: (err) => {
        reject(err)
      }
    })
  })
}

// 店铺
export const getShop = (shopId) => request({ url: `/shops/${shopId}` })

// 用户
export const getUserInfo = () => request({ url: '/auth/userinfo' })

export const loginByCode = (code) => request({
  url: '/auth/login',
  method: 'POST',
  data: { code },
  auth: false
})

// 钱包
export const getBalance = () => request({ url: '/wallet/balance' })

export const getWalletLogs = (page = 1, pageSize = 20) => request({
  url: `/wallet/logs?page=${page}&page_size=${pageSize}`
})

// 订单
export const getOrders = (page = 1, pageSize = 20) => request({
  url: `/orders?page=${page}&page_size=${pageSize}`
})

export const createOrder = (shopId, tableNo, totalAmount, items, useReward, orderType = 'dine_in') => request({
  url: '/orders',
  method: 'POST',
  data: {
    shop_id: shopId,
    order_type: orderType,
    table_no: tableNo,
    amount: totalAmount,
    items,
    use_reward: useReward
  }
})

// 邀请
export const getInviteStats = () => request({ url: '/invites/stats' })

export const bindInviteCode = (code, shopId) => request({ url: '/invites/bind', method: 'POST', data: { code, shop_id: shopId || 0 } })

export const getInviteQR = () => request({
  url: '/invites/qrcode',
  responseType: 'arraybuffer'
})

// 手机号验证
export const verifyPhone = (phone) => request({
  url: '/auth/verify-phone',
  method: 'POST',
  data: { phone }
})

// 福利金币
export const getRewardBalance = () => request({ url: '/reward/balance' })

export const getRewardLogs = (page = 1, pageSize = 20) => request({
  url: `/reward/logs?page=${page}&page_size=${pageSize}`
})

export const getRewardExpiryInfo = () => request({ url: '/reward/expiry-info' })

// 桌号绑定
export const getTableBinding = () => {
  return {
    shopId: wx.getStorageSync('current_shop_id') || 0,
    tableNo: wx.getStorageSync('current_table_no') || ''
  }
}

export const setTableBinding = (shopId, tableNo) => {
  wx.setStorageSync('current_shop_id', shopId)
  wx.setStorageSync('current_table_no', tableNo)
}