/**
 * Structural tests for the 我的 page aligning to mock-screens.html:
 * 粉渐变页头 + 黄头像白字 / 下划线式 tab / 金色余额卡（WCAG 折中）。
 * 金卡文字达标的颜色对比由 theme-tokens.test.js 计算把关。
 */
const fs = require('fs')
const path = require('path')

const wxss = fs.readFileSync(path.join(__dirname, '../pages/profile/index.wxss'), 'utf8')
const rule = (sel) => {
  const m = wxss.match(new RegExp('\\' + sel + '\\s*\\{([^}]*)\\}'))
  return m ? m[1] : null
}

describe('profile — 粉渐变页头（对齐设计稿）', () => {
  it('页头用品牌粉渐变，不再白底', () => {
    const h = rule('.user-header')
    expect(h).toMatch(/linear-gradient/)
    expect(h).toMatch(/var\(--weui-BRAND\)/)
    expect(h).not.toMatch(/background:\s*#fff/)
  })
  it('头像黄底粉字（设计稿字标头像）', () => {
    const a = rule('.avatar-placeholder')
    expect(a).toMatch(/background:\s*var\(--accent\)/)
    expect(a).toMatch(/color:\s*var\(--weui-BRAND\)/)
  })
  it('昵称/描述改白字（粉底上）', () => {
    expect(rule('.nickname')).toMatch(/color:\s*#fff/)
    expect(rule('.user-desc')).toMatch(/rgba\(255,\s*255,\s*255/)
  })
})

describe('profile — 下划线式 tab（对齐设计稿）', () => {
  it('激活态是粉字 + 粉下划线，不是粉底填充', () => {
    const on = rule('.weui-tabs__item_on')
    expect(on).toMatch(/color:\s*var\(--weui-BRAND\)/)
    expect(on).not.toMatch(/background:\s*var\(--weui-BRAND\)/)
    // 下划线条
    expect(wxss).toMatch(/\.weui-tabs__item_on::after\s*\{[^}]*background:\s*var\(--weui-BRAND\)/)
  })
})

describe('profile — 金色余额卡（WCAG 折中达标）', () => {
  it('卡面金色渐变 + 金描边', () => {
    const c = rule('.balance-card')
    expect(c).toMatch(/linear-gradient/)
    expect(c).toMatch(/var\(--jf-gold-bg-0\)/)
    expect(c).toMatch(/var\(--jf-gold-line\)/)
  })
  it('余额金额 + 标签用深琥珀字（达标）', () => {
    expect(wxss).toMatch(/\.balance-amount\.gold\s*\{[^}]*color:\s*var\(--jf-gold-ink\)/)
    expect(rule('.balance-label')).toMatch(/color:\s*var\(--jf-gold-ink\)/)
  })
})
