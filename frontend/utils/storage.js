/**
 * Storage 工具函数
 * 统一封装 wx.getStorageSync / wx.setStorageSync
 */

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

export const getToken = () => {
  return wx.getStorageSync('token') || ''
}

export const setToken = (token) => {
  wx.setStorageSync('token', token)
}

export const getUser = () => {
  return wx.getStorageSync('user') || null
}

export const setUser = (user) => {
  wx.setStorageSync('user', user)
}

export const getCart = (shopId) => {
  const key = `cart_${shopId}`
  return wx.getStorageSync(key) || []
}

export const setCart = (shopId, cart) => {
  const key = `cart_${shopId}`
  wx.setStorageSync(key, cart)
}

export const clearCart = (shopId) => {
  const key = `cart_${shopId}`
  wx.setStorageSync(key, [])
}

// 认证函数已迁移至 utils/auth.js，此处重导出以保持向后兼容
export { doLogin, isLoggedIn, requireLogin, handleAuthError } from './auth.js'