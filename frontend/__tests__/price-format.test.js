/**
 * formatPrice：菜单卡片主价格的展示格式。
 * 整数去掉无意义的 .00（¥100 而非 ¥100.00），非整价保留两位（¥38.50），
 * 让价格标更窄，配合紧凑按钮实现「价格 + 选规格」同行。
 */
const { formatPrice } = require('../utils/price.js')

describe('formatPrice', () => {
  it('整数去掉小数', () => {
    expect(formatPrice(100)).toBe('100')
    expect(formatPrice(38)).toBe('38')
    expect(formatPrice(1288)).toBe('1288')
    expect(formatPrice(0)).toBe('0')
  })

  it('非整数保留两位', () => {
    expect(formatPrice(38.5)).toBe('38.50')
    expect(formatPrice(38.55)).toBe('38.55')
    expect(formatPrice(38.1)).toBe('38.10')
  })

  it('字符串数字容错', () => {
    expect(formatPrice('100')).toBe('100')
    expect(formatPrice('38.5')).toBe('38.50')
  })

  it('非法输入回退为 0', () => {
    expect(formatPrice(undefined)).toBe('0')
    expect(formatPrice('abc')).toBe('0')
  })
})
