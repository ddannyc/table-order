// 规格(SKU)草稿落库的纯逻辑：把「编辑弹层草稿」与「服务端原始规格」算成增/改/删意图，
// 供保存时驱动既有规格接口（createProductSpec/updateProductSpec/deleteProductSpec）。
// 不发请求、无副作用，便于单测。规格为扁平单维 {id?, name, price, status}。

const trimmed = (v) => String(v == null ? '' : v).trim()

// 同一规格是否有实质变化（name/price/status 任一不同）。price 容忍字符串↔数字等值。
function specChanged(a, b) {
  return (
    trimmed(a.name) !== trimmed(b.name) ||
    Number(a.price) !== Number(b.price) ||
    Number(a.status) !== Number(b.status)
  )
}

// 草稿 vs 原始 → { creates, updates, deletes }
//  creates: 无 id 的草稿行（仅 name/price/status）
//  updates: 有 id 且相对原始有变化的行（id + name/price/status）
//  deletes: 原始存在、草稿已移除的 id
export function diffSpecs(original, draft) {
  const orig = Array.isArray(original) ? original : []
  const rows = Array.isArray(draft) ? draft : []
  const origById = new Map(orig.map((s) => [s.id, s]))

  const creates = []
  const updates = []
  for (const r of rows) {
    if (!r.id) {
      creates.push({ name: r.name, price: Number(r.price), status: Number(r.status) })
      continue
    }
    const o = origById.get(r.id)
    if (o && specChanged(o, r)) {
      updates.push({ id: r.id, name: r.name, price: Number(r.price), status: Number(r.status) })
    }
  }

  const draftIds = new Set(rows.filter((r) => r.id).map((r) => r.id))
  const deletes = orig.filter((o) => !draftIds.has(o.id)).map((o) => o.id)

  return { creates, updates, deletes }
}

// 落库前校验：空草稿合法；每行 name 非空、price>0。
export function validateSpecs(draft) {
  const rows = Array.isArray(draft) ? draft : []
  for (const r of rows) {
    if (!trimmed(r.name)) {
      return { ok: false, message: '规格名称不能为空' }
    }
    const price = Number(r.price)
    if (!(price > 0)) {
      return { ok: false, message: '规格价格必须大于 0' }
    }
  }
  return { ok: true }
}
