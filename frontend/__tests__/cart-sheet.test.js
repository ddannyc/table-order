/**
 * 购物车弹层：行展示纯逻辑（utils/cart-view.js）守护行数据/价格/小计；
 * 改量/清空走既有 updateCartQuantity/clearCart（由 cart-sku/cart-isolation 守护）。
 * 弹层版式（T4）由本文件后续结构断言守护。
 */
const fs = require('fs')
const path = require('path')
const { buildCartItems } = require('../utils/cart-view.js')

const wxml = fs.readFileSync(path.join(__dirname, '../pages/menu/index.wxml'), 'utf8')
const wxss = fs.readFileSync(path.join(__dirname, '../pages/menu/index.wxss'), 'utf8')
const rule = (sel) => {
  const m = wxss.match(new RegExp('\\' + sel + '\\s*\\{([^}]*)\\}'))
  return m ? m[1] : null
}

const cart = [
  { key: '1_5', productId: 1, name: '盐焗鸡', specName: '中份', price: 25, image: 'a.png', quantity: 2 },
  { key: '3_0', productId: 3, name: '可乐', specName: '', price: 6, image: '', quantity: 1 },
]

describe('buildCartItems — 购物车行展示态', () => {
  it('每行带单价文案与小计文案', () => {
    const rows = buildCartItems(cart)
    expect(rows[0].priceText).toBe('25.00')
    expect(rows[0].lineTotalText).toBe('50.00') // 25 * 2
    expect(rows[1].priceText).toBe('6.00')
    expect(rows[1].lineTotalText).toBe('6.00')
  })
  it('保留 key/名称/规格/数量/缩略图', () => {
    const rows = buildCartItems(cart)
    expect(rows[0]).toMatchObject({ key: '1_5', name: '盐焗鸡', specName: '中份', quantity: 2, image: 'a.png' })
  })
  it('空/非数组返回空数组', () => {
    expect(buildCartItems([])).toEqual([])
    expect(buildCartItems(undefined)).toEqual([])
  })
})

describe('购物车弹层版式（已选商品面板 + JFW）', () => {
  it('遮罩 + 底部 sheet，遮罩点击关层', () => {
    expect(wxml).toMatch(/class="cart-mask"[^>]*bindtap="closeCartSheet"/)
    expect(wxml).toMatch(/class="cart-sheet"/)
  })
  it('头部「已选商品」+ 清空（clearCartAll）', () => {
    expect(wxml).toMatch(/已选商品/)
    expect(wxml).toMatch(/清空/)
    expect(wxml).toMatch(/bindtap="clearCartAll"/)
  })
  it('行循环 cartItems，每行带 −/＋ 步进（cartDec/cartInc + data-key）', () => {
    expect(wxml).toMatch(/wx:for="\{\{cartItems\}\}"/)
    expect(wxml).toMatch(/bindtap="cartDec"\s+data-key/)
    expect(wxml).toMatch(/bindtap="cartInc"\s+data-key/)
  })
  it('弹层内可去结算（goCart）', () => {
    const sheet = wxml.slice(wxml.indexOf('cart-sheet'))
    expect(sheet).toMatch(/bindtap="goCart"/)
  })
  it('行价用圆体数字令牌、步进沿用粉边圆钮', () => {
    expect(rule('.cart-row-price')).toMatch(/font-family:\s*var\(--font-number\)/)
    expect(wxss).toMatch(/\.cart-sheet/)
  })
})
