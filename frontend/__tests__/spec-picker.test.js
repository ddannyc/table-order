/**
 * 规格弹层纯逻辑（utils/spec.js）。Page 方法难直接单测，故把判定/计算抽成纯函数在这里守护；
 * 加购行为由 cart-sku.test.js 守护（addToCart 不变），弹层版式由本文件后续结构断言守护。
 */
const { pickDefaultSpec, clampQty, specPickerState } = require('../utils/spec.js')

const specs = [
  { id: 5, name: '小份', price: 18, status: 1 },
  { id: 6, name: '中份', price: 25, status: 1 },
  { id: 7, name: '家庭桶', price: 58, status: 2 }, // 售罄
]

describe('pickDefaultSpec — 默认选首个在售规格', () => {
  it('返回第一个 status===1 的规格', () => {
    expect(pickDefaultSpec(specs).id).toBe(5)
  })
  it('跳过售罄，选到第一个在售', () => {
    expect(pickDefaultSpec([{ id: 1, status: 2 }, { id: 2, status: 1 }]).id).toBe(2)
  })
  it('全售罄返回 null', () => {
    expect(pickDefaultSpec([{ id: 1, status: 2 }, { id: 2, status: 0 }])).toBeNull()
  })
  it('空/非数组返回 null', () => {
    expect(pickDefaultSpec([])).toBeNull()
    expect(pickDefaultSpec(undefined)).toBeNull()
  })
})

describe('clampQty — 数量下限 1、取整', () => {
  it('低于 1 夹到 1', () => {
    expect(clampQty(0)).toBe(1)
    expect(clampQty(-3)).toBe(1)
  })
  it('正常值原样', () => {
    expect(clampQty(2)).toBe(2)
  })
  it('取整、非法回退 1', () => {
    expect(clampQty(3.7)).toBe(3)
    expect(clampQty(NaN)).toBe(1)
    expect(clampQty(undefined)).toBe(1)
  })
})

describe('specPickerState — 弹层展示态', () => {
  it('选中在售规格：单价/合计/可加入', () => {
    const st = specPickerState(specs, 6, 2)
    expect(st.selectedSpecId).toBe(6)
    expect(st.unitText).toBe('25.00')
    expect(st.totalText).toBe('50.00') // 25 * 2
    expect(st.canAdd).toBe(true)
  })
  it('数量夹到下限再算合计', () => {
    expect(specPickerState(specs, 5, 0).totalText).toBe('18.00') // qty->1
  })
  it('选中售罄：不可加入，合计 0', () => {
    const st = specPickerState(specs, 7, 1)
    expect(st.canAdd).toBe(false)
    expect(st.totalText).toBe('0.00')
  })
  it('无选中（null/不存在）：不可加入', () => {
    expect(specPickerState(specs, null, 1).canAdd).toBe(false)
    expect(specPickerState(specs, 999, 1).canAdd).toBe(false)
  })
})
