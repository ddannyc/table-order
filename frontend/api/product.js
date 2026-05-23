/**
 * 购物车 / 产品 API
 */

import { getCart, setCart, clearCart } from '../utils/storage.js'

const { API_BASE } = require('../config.js')

export { getCart, clearCart }

export const getShopProducts = (shopId) => {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE}/shops/${shopId}/products`,
      method: 'GET',
      success: (res) => resolve(res.data),
      fail: reject
    })
  })
}

export const getCartTotal = (shopId) => {
  const cart = getCart(shopId)
  return cart.reduce((sum, item) => sum + item.price * item.quantity, 0)
}

export const getCartCount = (shopId) => {
  const cart = getCart(shopId)
  return cart.reduce((sum, item) => sum + item.quantity, 0)
}

export const addToCart = (shopId, product, quantity = 1) => {
  const cart = getCart(shopId)
  const existing = cart.find(item => item.id === product.id)
  if (existing) {
    existing.quantity += quantity
  } else {
    cart.push({ ...product, quantity })
  }
  setCart(shopId, cart)
  return cart
}

export const updateCartQuantity = (shopId, productId, quantity) => {
  const cart = getCart(shopId)
  if (quantity <= 0) {
    const filtered = cart.filter(item => item.id !== productId)
    setCart(shopId, filtered)
  } else {
    const item = cart.find(item => item.id === productId)
    if (item) {
      item.quantity = quantity
      setCart(shopId, cart)
    }
  }
  return cart
}