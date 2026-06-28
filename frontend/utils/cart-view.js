/**
 * 购物车弹层行展示纯逻辑（可单测）。把存储里的购物车项映射成弹层渲染所需字段。
 * 改量/删行/清空仍走 api/product 的 updateCartQuantity/clearCart。
 */

// 购物车项 -> 弹层行：补单价文案与小计文案，保留 key/名称/规格/数量/缩略图
function buildCartItems(cart) {
  if (!Array.isArray(cart)) return []
  return cart.map((i) => ({
    key: i.key,
    productId: i.productId,
    name: i.name,
    specName: i.specName || '',
    image: i.image || '',
    quantity: i.quantity,
    priceText: i.price.toFixed(2),
    lineTotalText: (i.price * i.quantity).toFixed(2),
  }))
}

module.exports = { buildCartItems }
