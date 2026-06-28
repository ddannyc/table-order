import { describe, it, expect } from 'vitest'
import { diffSpecs, validateSpecs, toDraftSpecs } from './specSync'

const orig = [
  { id: 1, name: '小份', price: 18, status: 1 },
  { id: 2, name: '中份', price: 25, status: 1 },
  { id: 3, name: '大份', price: 32, status: 2 },
]

describe('diffSpecs — 草稿 vs 原始 → 落库意图', () => {
  it('无 id 的草稿行 → creates', () => {
    const draft = [...orig, { name: '家庭桶', price: 58, status: 1 }]
    const d = diffSpecs(orig, draft)
    expect(d.creates).toEqual([{ name: '家庭桶', price: 58, status: 1 }])
    expect(d.updates).toEqual([])
    expect(d.deletes).toEqual([])
  })

  it('改价/改名/改状态的已存在行 → updates（仅变化项）', () => {
    const draft = [
      { id: 1, name: '小份', price: 20, status: 1 }, // 改价
      { id: 2, name: '中份', price: 25, status: 1 }, // 不变
      { id: 3, name: '大份', price: 32, status: 1 }, // 改状态
    ]
    const d = diffSpecs(orig, draft)
    expect(d.updates).toEqual([
      { id: 1, name: '小份', price: 20, status: 1 },
      { id: 3, name: '大份', price: 32, status: 1 },
    ])
    expect(d.creates).toEqual([])
    expect(d.deletes).toEqual([])
  })

  it('原始有而草稿缺的 id → deletes', () => {
    const draft = [
      { id: 1, name: '小份', price: 18, status: 1 },
      { id: 3, name: '大份', price: 32, status: 2 },
    ]
    expect(diffSpecs(orig, draft).deletes).toEqual([2])
  })

  it('完全未变 → 三类皆空', () => {
    const draft = orig.map((s) => ({ ...s }))
    const d = diffSpecs(orig, draft)
    expect(d).toEqual({ creates: [], updates: [], deletes: [] })
  })

  it('价格字符串与数字等值不算变化（表单可能给字符串）', () => {
    const draft = [
      { id: 1, name: '小份', price: '18', status: 1 },
      { id: 2, name: '中份', price: 25, status: 1 },
      { id: 3, name: '大份', price: 32, status: 2 },
    ]
    expect(diffSpecs(orig, draft).updates).toEqual([])
  })

  it('增删改混合', () => {
    const draft = [
      { id: 1, name: '小份', price: 19, status: 1 }, // update
      // id 2 删除
      { id: 3, name: '大份', price: 32, status: 2 }, // 不变
      { name: '超大份', price: 40, status: 1 }, // create
    ]
    const d = diffSpecs(orig, draft)
    expect(d.creates).toEqual([{ name: '超大份', price: 40, status: 1 }])
    expect(d.updates).toEqual([{ id: 1, name: '小份', price: 19, status: 1 }])
    expect(d.deletes).toEqual([2])
  })

  it('空草稿 → 全删', () => {
    expect(diffSpecs(orig, []).deletes).toEqual([1, 2, 3])
  })

  it('原始为空、草稿全新 → 全 creates', () => {
    const d = diffSpecs([], [{ name: 'A', price: 5, status: 1 }])
    expect(d.creates).toEqual([{ name: 'A', price: 5, status: 1 }])
    expect(d.updates).toEqual([])
    expect(d.deletes).toEqual([])
  })
})

describe('validateSpecs — 落库前校验', () => {
  it('空草稿合法（无规格按菜品价售卖）', () => {
    expect(validateSpecs([]).ok).toBe(true)
  })

  it('全部有效 → ok', () => {
    expect(validateSpecs(orig).ok).toBe(true)
  })

  it('空名（含纯空格）→ 不通过', () => {
    expect(validateSpecs([{ name: '  ', price: 10, status: 1 }]).ok).toBe(false)
    expect(validateSpecs([{ name: '', price: 10, status: 1 }]).ok).toBe(false)
  })

  it('价格 <=0 或非数 → 不通过', () => {
    expect(validateSpecs([{ name: 'A', price: 0, status: 1 }]).ok).toBe(false)
    expect(validateSpecs([{ name: 'A', price: -1, status: 1 }]).ok).toBe(false)
    expect(validateSpecs([{ name: 'A', price: 'x', status: 1 }]).ok).toBe(false)
  })

  it('不通过时带可读 message', () => {
    expect(typeof validateSpecs([{ name: '', price: 1, status: 1 }]).message).toBe('string')
  })
})

describe('toDraftSpecs — 服务端规格 → 可编辑草稿', () => {
  it('只保留 id/name/price/status，price/status 归一为数字', () => {
    const server = [{ id: 5, name: '中份', price: '25.00', status: 1, created_at: 'x', product_id: 9 }]
    expect(toDraftSpecs(server)).toEqual([{ id: 5, name: '中份', price: 25, status: 1 }])
  })

  it('深拷贝：改草稿不影响源对象（保护表格与 diff 基线）', () => {
    const server = [{ id: 1, name: '小份', price: 18, status: 1 }]
    const draft = toDraftSpecs(server)
    draft[0].price = 99
    draft[0].name = '改了'
    expect(server[0].price).toBe(18)
    expect(server[0].name).toBe('小份')
  })

  it('空/未定义 → []', () => {
    expect(toDraftSpecs([])).toEqual([])
    expect(toDraftSpecs(undefined)).toEqual([])
  })
})
