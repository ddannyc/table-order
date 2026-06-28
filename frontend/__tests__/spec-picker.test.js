/**
 * 规格弹层纯逻辑（utils/spec.js）。Page 方法难直接单测，故把判定/计算抽成纯函数在这里守护；
 * 加购行为由 cart-sku.test.js 守护（addToCart 不变），弹层版式由本文件后续结构断言守护。
 */
const fs = require('fs')
const path = require('path')
const { pickDefaultSpec, clampQty, specPickerState } = require('../utils/spec.js')

const wxml = fs.readFileSync(path.join(__dirname, '../pages/menu/index.wxml'), 'utf8')
const wxss = fs.readFileSync(path.join(__dirname, '../pages/menu/index.wxss'), 'utf8')
const rule = (sel) => {
  const m = wxss.match(new RegExp('\\' + sel + '\\s*\\{([^}]*)\\}'))
  return m ? m[1] : null
}

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

describe('规格弹层版式（对齐参考图 + JFW）', () => {
  it('单选药丸组：绑定 selectSpec，按选中/售罄切类', () => {
    expect(wxml).toMatch(/class="spec-pill[^"]*spec-pill_on/)
    expect(wxml).toMatch(/spec-pill_off/)
    expect(wxml).toMatch(/bindtap="selectSpec"\s+data-spec-id/)
  })
  it('数量步进绑定 specDec/specInc', () => {
    expect(wxml).toMatch(/bindtap="specDec"/)
    expect(wxml).toMatch(/bindtap="specInc"/)
  })
  it('整宽加入按钮：weui-btn_primary + bindtap confirmAddSpec，无 type="primary"', () => {
    expect(wxml).toMatch(/class="[^"]*weui-btn_primary[^"]*spec-add[^"]*"[^>]*bindtap="confirmAddSpec"/)
    // weui-btn 默认收缩居中，须显式整宽
    expect(rule('.spec-add')).toMatch(/width:\s*100%/)
    expect(wxml).not.toMatch(/pickSpec/)
    // 规格弹层内不得出现原生 type="primary"
    const sheet = wxml.slice(wxml.indexOf('spec-sheet'))
    expect(sheet).not.toMatch(/type="primary"/)
  })
  it('选中药丸品牌粉、价格用圆体数字令牌', () => {
    expect(rule('.spec-pill_on')).toMatch(/background:\s*var\(--weui-BRAND\)/)
    expect(rule('.spec-price')).toMatch(/font-family:\s*var\(--font-number\)/)
  })
  it('板块标题深蓝 + 粉 #（与菜单一致）', () => {
    expect(rule('.spec-group-t')).toMatch(/color:\s*var\(--jf-title-blue\)/)
    expect(wxss).toMatch(/\.spec-group-t::before\s*\{[^}]*content:\s*"#"/)
  })
})

describe('规格弹层头图（变体 A · 整宽封面图）', () => {
  it('移除顶部短横线 grab', () => {
    expect(wxml).not.toMatch(/spec-grab/)
    expect(wxss).not.toMatch(/\.spec-grab/)
  })
  it('标题上方有整宽封面图（真实图 + 占位两路），关闭✕落在封面上', () => {
    expect(wxml).toMatch(/class="spec-cover"/)
    const sheet = wxml.slice(wxml.indexOf('spec-sheet'))
    // 封面排在标题之前
    expect(sheet.indexOf('spec-cover')).toBeLessThan(sheet.indexOf('spec-name'))
    // 有图走 image，无图回退占位块
    expect(wxml).toMatch(/class="spec-cover-img"[^>]*src="\{\{specPickerProduct\.image\}\}"/)
    expect(wxml).toMatch(/spec-cover-ph/)
    // ✕ 浮在封面上、绑 closeSpecPicker
    expect(wxml).toMatch(/spec-cover-x[^>]*bindtap="closeSpecPicker"/)
  })
  it('封面整宽铺满 sheet（负边距抵消 padding）', () => {
    expect(rule('.spec-cover')).toMatch(/margin/)
  })
  it('弹层盖在固定 tabbar(z-index:100) 之上，加入购物车按钮不被遮挡', () => {
    const z = (sel) => Number((String(rule(sel)).match(/z-index:\s*(\d+)/) || [])[1])
    expect(z('.spec-mask')).toBeGreaterThan(100)
    expect(z('.spec-sheet')).toBeGreaterThan(100)
  })
})
