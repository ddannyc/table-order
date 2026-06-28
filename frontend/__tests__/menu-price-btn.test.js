/**
 * 菜单卡片「选规格」按钮在三位数价格下不再被挤成两行（方案 V3）。
 * - 按钮带专属类 menu-spec-btn，且不换行（white-space:nowrap）、不收缩（flex-shrink:0）。
 * - 卡片底行可兜底换行（flex-wrap:wrap）：极端超宽时整颗按钮掉次行，绝不拆字。
 */
const fs = require('fs')
const path = require('path')
const read = (p) => fs.readFileSync(path.join(__dirname, '..', p), 'utf8')

describe('菜单「选规格」按钮不被三位数价格挤断行', () => {
  const wxml = read('pages/menu/index.wxml')
  const wxss = read('pages/menu/index.wxss')

  it('选规格按钮带 menu-spec-btn 类', () => {
    const btnLine = wxml.split('\n').find((l) => l.includes('openSpecPicker'))
    expect(btnLine).toBeTruthy()
    expect(btnLine).toMatch(/menu-spec-btn/)
  })

  it('menu-spec-btn 不换行、不收缩', () => {
    const m = wxss.match(/\.menu-spec-btn\s*\{([^}]*)\}/)
    expect(m).toBeTruthy()
    expect(m[1]).toMatch(/white-space:\s*nowrap/)
    expect(m[1]).toMatch(/flex-shrink:\s*0/)
  })

  it('卡片底行可兜底换行', () => {
    const m = wxss.match(/\.menu-card-bottom\s*\{([^}]*)\}/)
    expect(m).toBeTruthy()
    expect(m[1]).toMatch(/flex-wrap:\s*wrap/)
  })
})
