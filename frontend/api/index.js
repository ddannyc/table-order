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

// 外卖：解析配送门店（当前单门店）
export const resolveDeliveryShop = () => request({ url: '/delivery/shop' })

// 外卖：实时运费报价（返回 delivery_fee + 后端签名报价凭证 quote_token）
export const getDeliveryQuote = (shopId, recipient) => request({
  url: '/delivery/quote',
  method: 'POST',
  data: {
    shop_id: shopId,
    recipient_name: recipient.recipient_name,
    recipient_phone: recipient.recipient_phone,
    recipient_address: recipient.recipient_address,
    recipient_lat: recipient.lat,
    recipient_lng: recipient.lng
  }
})

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

export const createOrder = (shopId, tableNo, totalAmount, items, useReward, orderType = 'dine_in', delivery = null, quoteToken = '') => request({
  url: '/orders',
  method: 'POST',
  data: {
    shop_id: shopId,
    order_type: orderType,
    table_no: tableNo,
    amount: totalAmount,
    items,
    use_reward: useReward,
    // delivery fields only for delivery orders
    ...(orderType === 'delivery' ? { delivery, quote_token: quoteToken } : {})
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