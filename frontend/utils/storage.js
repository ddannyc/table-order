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

export const getLastDeliveryAddress = () => {
  return wx.getStorageSync('last_delivery_address') || null
}

export const setLastDeliveryAddress = (addr) => {
  wx.setStorageSync('last_delivery_address', addr)
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

// Cart storage is keyed by shop AND order type, so a shop's dine-in and
// delivery carts never share storage (orderType defaults to dine_in).
const cartStoreKey = (shopId, orderType = 'dine_in') => `cart_v2_${shopId}_${orderType}`

export const getCart = (shopId, orderType = 'dine_in') => {
  return wx.getStorageSync(cartStoreKey(shopId, orderType)) || []
}

export const setCart = (shopId, cart, orderType = 'dine_in') => {
  wx.setStorageSync(cartStoreKey(shopId, orderType), cart)
}

export const clearCart = (shopId, orderType = 'dine_in') => {
  wx.setStorageSync(cartStoreKey(shopId, orderType), [])
}

// 认证函数已迁移至 utils/auth.js，此处重导出以保持向后兼容
export { doLogin, isLoggedIn, requireLogin, handleAuthError } from './auth.js'