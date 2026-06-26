/**
 * 菜品缩略图解析（M3）
 * 不引入二进制资源：无图菜品按分类回退到 CSS 占位块描述符（中性底 + 字形/标签，
 * 配色保持中性，把唯一亮点留给左侧类目轨）。有图则用真实图。
 */

// 占位描述符：label 用于占位块文字，glyph 供 CSS 选择不同字形/底纹。
const CATEGORY_PLACEHOLDERS = {
  新品上市: { label: '新品', glyph: 'sparkle' },
  芝士奶盖: { label: '芝士', glyph: 'cheese' },
  奶茶牛乳: { label: '奶茶', glyph: 'cup' },
  气泡水: { label: '气泡', glyph: 'bubble' },
}

const DEFAULT_PLACEHOLDER = { label: '饮品', glyph: 'cup' }

const placeholderFor = (category) => CATEGORY_PLACEHOLDERS[category] || DEFAULT_PLACEHOLDER

// 解析一个菜品的缩略图渲染方式：有 image 用真实图，否则回退到分类占位块。
const resolveProductImage = (product) => {
  const placeholder = placeholderFor(product && product.category)
  const hasImage = !!(product && product.image)
  return { hasImage, src: hasImage ? product.image : '', placeholder }
}

module.exports = { placeholderFor, resolveProductImage }
