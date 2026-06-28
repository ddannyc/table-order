// utils/price.js — 价格展示格式化
// 整数去掉无意义的 .00（¥100），非整价保留两位（¥38.50）。容错非法入参回退 0。
function formatPrice(n) {
  const num = Number(n) || 0
  return Number.isInteger(num) ? String(num) : num.toFixed(2)
}

module.exports = { formatPrice }
