/**
 * Tests for SKU-aware cart (Task 10).
 * Cart lines are keyed by productId_specId; spec lines use the spec price;
 * no-spec products use specId 0 and the product price.
 */
let store = {}
global.wx = {
  getAccountInfoSync: () => ({ miniProgram: { envVersion: 'develop' } }),
  getStorageSync: (k) => store[k],
  setStorageSync: (k, v) => {
    store[k] = v
  },
}

const { addToCart, updateCartQuantity, getCart, getCartCount, getCartTotal } = require('../api/product.js')

const product = { id: 10, name: '酸奶青提', price: 15, image: '' }
const spec800 = { id: 5, name: '800ml', price: 18 }
const spec600 = { id: 6, name: '600ml', price: 15 }

beforeEach(() => {
  store = {}
})

describe('addToCart with specs', () => {
  it('creates a composite-key line priced at the spec price', () => {
    addToCart(1, product, spec800, 1)
    const cart = getCart(1)
    expect(cart).toHaveLength(1)
    expect(cart[0].key).toBe('10_5')
    expect(cart[0].price).toBe(18)
    expect(cart[0].specName).toBe('800ml')
    expect(cart[0].specId).toBe(5)
  })

  it('keeps different specs of the same product as separate lines', () => {
    addToCart(1, product, spec800, 1)
    addToCart(1, product, spec600, 1)
    expect(getCart(1)).toHaveLength(2)
  })

  it('increments quantity for the same product+spec', () => {
    addToCart(1, product, spec800, 1)
    addToCart(1, product, spec800, 1)
    const cart = getCart(1)
    expect(cart).toHaveLength(1)
    expect(cart[0].quantity).toBe(2)
  })

  it('uses specId 0 and product price for a no-spec product', () => {
    addToCart(1, product, null, 1)
    const cart = getCart(1)
    expect(cart[0].key).toBe('10_0')
    expect(cart[0].specId).toBe(0)
    expect(cart[0].price).toBe(15)
  })
})

describe('updateCartQuantity by key', () => {
  it('removes the line at quantity 0', () => {
    addToCart(1, product, spec800, 1)
    updateCartQuantity(1, '10_5', 0)
    expect(getCart(1)).toHaveLength(0)
  })

  it('sets the quantity for a key', () => {
    addToCart(1, product, spec800, 1)
    updateCartQuantity(1, '10_5', 3)
    expect(getCart(1)[0].quantity).toBe(3)
  })
})

describe('totals use spec price', () => {
  it('sums spec price * quantity', () => {
    addToCart(1, product, spec800, 2)
    expect(getCartTotal(1)).toBe(36)
    expect(getCartCount(1)).toBe(2)
  })
})
