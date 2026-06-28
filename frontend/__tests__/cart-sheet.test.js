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
