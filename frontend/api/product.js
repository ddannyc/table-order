/**
 * 购物车 / 产品 API
 */

import { getCart, setCart, clearCart } from '../utils/storage.js'

const { request } = require('../api/index.js')

export { getCart, clearCart }

export const getShopProducts = (shopId) => {
  return request({ url: `/shops/${shopId}/products` })
}

export const getCartTotal = (shopId) => {
  const cart = getCart(shopId)
  return cart.reduce((sum, item) => sum + item.price * item.quantity, 0)
}

export const getCartCount = (shopId) => {
  const cart = getCart(shopId)
  return cart.reduce((sum, item) => sum + item.quantity, 0)
}

// Cart lines are keyed by productId_specId (specId 0 = no spec).
const cartKey = (productId, specId) => `${productId}_${specId || 0}`

export const addToCart = (shopId, product, spec, quantity = 1) => {
  const cart = getCart(shopId)
  const specId = spec ? spec.id : 0
  const key = cartKey(product.id, specId)
  const existing = cart.find(item => item.key === key)
  if (existing) {
    existing.quantity += quantity
  } else {
    cart.push({
      key,
      productId: product.id,
      specId,
      name: product.name,
      specName: spec ? spec.name : '',
      price: spec ? spec.price : product.price,
      image: product.image,
      quantity
    })
  }
  setCart(shopId, cart)
  return cart
}

export const updateCartQuantity = (shopId, key, quantity) => {
  const cart = getCart(shopId)
  if (quantity <= 0) {
    setCart(shopId, cart.filter(item => item.key !== key))
  } else {
    const item = cart.find(item => item.key === key)
    if (item) {
      item.quantity = quantity
      setCart(shopId, cart)
    }
  }
  return cart
}