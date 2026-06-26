/**
 * Cart isolation by order type (M2).
 * Dine-in and delivery carts for the SAME shop must not share storage, so a
 * customer switching mode via the home entry never clobbers the other cart.
 */
let store = {}
global.wx = {
  getAccountInfoSync: () => ({ miniProgram: { envVersion: 'develop' } }),
  getStorageSync: (k) => store[k],
  setStorageSync: (k, v) => {
    store[k] = v
  },
}

const { addToCart, getCart, getCartCount, getCartTotal, clearCart } = require('../api/product.js')

const product = { id: 10, name: '招牌奶茶', price: 13, image: '' }

beforeEach(() => {
  store = {}
})

describe('cart isolation by order type', () => {
  it('keeps dine_in and delivery carts independent for the same shop', () => {
    addToCart(5, product, null, 1, 'dine_in')
    addToCart(5, product, null, 2, 'delivery')

    expect(getCartCount(5, 'dine_in')).toBe(1)
    expect(getCartCount(5, 'delivery')).toBe(2)
    expect(getCartTotal(5, 'dine_in')).toBe(13)
    expect(getCartTotal(5, 'delivery')).toBe(26)
    expect(getCart(5, 'dine_in')).toHaveLength(1)
    expect(getCart(5, 'delivery')).toHaveLength(1)
  })

  it('clearing one mode does not affect the other', () => {
    addToCart(5, product, null, 1, 'dine_in')
    addToCart(5, product, null, 2, 'delivery')

    clearCart(5, 'dine_in')

    expect(getCartCount(5, 'dine_in')).toBe(0)
    expect(getCartCount(5, 'delivery')).toBe(2)
  })
})
