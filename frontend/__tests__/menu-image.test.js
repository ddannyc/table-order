/**
 * Menu image resolution (M3).
 * No binary assets: products without an image fall back to a per-category CSS
 * placeholder descriptor; products with an image use it.
 */
const { placeholderFor, resolveProductImage } = require('../utils/menu-image.js')

describe('placeholderFor', () => {
  it('returns a category-specific placeholder for known categories', () => {
    expect(placeholderFor('奶茶牛乳').label).toBe('奶茶')
    expect(placeholderFor('芝士奶盖').label).toBe('芝士')
    expect(placeholderFor('气泡水').label).toBe('气泡')
    expect(placeholderFor('新品上市').label).toBe('新品')
  })

  it('falls back to a generic placeholder for unknown/empty categories', () => {
    expect(placeholderFor('不存在分类').label).toBe('饮品')
    expect(placeholderFor(undefined).label).toBe('饮品')
  })
})

describe('resolveProductImage', () => {
  it('uses the real image when present', () => {
    const r = resolveProductImage({ image: 'https://x/a.png', category: '奶茶牛乳' })
    expect(r.hasImage).toBe(true)
    expect(r.src).toBe('https://x/a.png')
  })

  it('falls back to a category placeholder when no image', () => {
    const r = resolveProductImage({ image: '', category: '气泡水' })
    expect(r.hasImage).toBe(false)
    expect(r.src).toBe('')
    expect(r.placeholder.label).toBe('气泡')
  })

  it('is null-safe', () => {
    const r = resolveProductImage(undefined)
    expect(r.hasImage).toBe(false)
    expect(r.placeholder.label).toBe('饮品')
  })
})
