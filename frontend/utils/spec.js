/**
 * 规格弹层纯逻辑（可单测）。供 pages/menu 规格选择层使用。
 * 规格为扁平单维（一个规格=一个价），status: 1=在售 0=下架 2=售罄。
 */

// 默认选中第一个在售规格；无在售返回 null
function pickDefaultSpec(specs) {
  if (!Array.isArray(specs)) return null
  return specs.find((s) => s && s.status === 1) || null
}

// 步进数量下限 1、取整、非法回退 1
function clampQty(n, min = 1) {
  const v = Math.floor(Number(n))
  if (!Number.isFinite(v) || v < min) return min
  return v
}

// 弹层展示态：选中规格、单价、合计、能否加入
function specPickerState(specs, selectedSpecId, qty) {
  const list = Array.isArray(specs) ? specs : []
  const selected = list.find((s) => String(s.id) === String(selectedSpecId)) || null
  const sellable = !!(selected && selected.status === 1)
  const q = clampQty(qty)
  const unit = sellable ? selected.price : 0
  return {
    selectedSpecId: selected ? selected.id : null,
    unitText: (selected ? selected.price : 0).toFixed(2),
    totalText: (unit * q).toFixed(2),
    canAdd: sellable,
  }
}

module.exports = { pickDefaultSpec, clampQty, specPickerState }
