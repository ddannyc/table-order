/**
 * Tests for the 鸡福旺 (JFW) brand design tokens.
 * Guards three things:
 *   1. Each token is defined in app.wxss with its agreed hex value.
 *   2. .page paints the page background (--weui-BG-0).
 *   3. The chosen colors meet WCAG contrast floors — computed here, so a
 *      future token edit that quietly breaks contrast fails the suite.
 *
 * WCAG policy (see tasks/plan.md 决策 #3): small text keeps AA 4.5:1; the bright
 * brand pink (--weui-BRAND) carries only fills + large brand text (AA-large 3:1,
 * the correct threshold for >=18px/bold). Small light text rides --green-2 (deep
 * pink). The price tag is red text on a light card, not red-on-yellow.
 */
const fs = require('fs')
const path = require('path')

const css = fs.readFileSync(path.join(__dirname, '../app.wxss'), 'utf8')

const token = (name) => {
  const m = css.match(new RegExp(name + '\\s*:\\s*(#[0-9A-Fa-f]{6})'))
  return m ? m[1].toUpperCase() : null
}

// --- WCAG 2.x relative luminance + contrast ratio ---
const channel = (c) => {
  const s = c / 255
  return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4)
}
const luminance = (hex) => {
  const [r, g, b] = hex.replace('#', '').match(/.{2}/g).map((h) => parseInt(h, 16))
  return 0.2126 * channel(r) + 0.7152 * channel(g) + 0.0722 * channel(b)
}
const contrast = (a, b) => {
  const la = luminance(a)
  const lb = luminance(b)
  const [hi, lo] = la > lb ? [la, lb] : [lb, la]
  return (hi + 0.05) / (lo + 0.05)
}

describe('theme tokens — 鸡福旺 (JFW) palette', () => {
  const expected = {
    '--weui-BRAND': '#FF4896', // 主粉：按钮/激活/cartbar/徽章
    '--green-2': '#D81A60', // 深粉：渐变深端 + 承载小号浅色文字（白字 ≈4.96:1）
    '--weui-BG-0': '#FBEFF3', // 页面·淡粉
    '--weui-BG-1': '#FFFDF8', // 卡面·暖近白
    '--weui-BG-2': '#FCE7EF', // 左类目轨底
    '--weui-FG-0': '#222222', // 正文
    '--weui-FG-2': '#6E646A', // 次要文字（>=4.5:1 on 页/卡）
    '--weui-FG-3': '#E8E8E8', // 分割线
    '--accent': '#FFD300', // 亮黄：优惠圈/装饰（不垫价格数字）
    '--price-ink': '#D11414', // 价格红（on 卡 ≈5.4:1）
    '--jf-blue': '#0066FF', // 蓝点缀
    '--jf-title-blue': '#0A2463', // # 板块标题
    '--jf-tag-pink': '#FFE0ED', // 优惠圈底
    '--jf-tag-red': '#C2185B', // 优惠圈字
    '--jf-orange': '#FFB829', // 饮品/果茶
    '--jf-gold-bg-0': '#FFF4DA', // 金卡渐变浅端
    '--jf-gold-bg-1': '#FFE7A8', // 金卡渐变深端
    '--jf-gold-line': '#FFE08A', // 金卡描边
    '--jf-gold-ink': '#8A5500', // 金卡深琥珀字
  }

  it.each(Object.entries(expected))('defines %s = %s', (name, hex) => {
    expect(token(name)).toBe(hex)
  })

  it('.page paints the page background (--weui-BG-0)', () => {
    expect(css).toMatch(/\.page\s*\{[^}]*background:\s*var\(--weui-BG-0\)/)
  })

  // weui declares --weui-BRAND:#07c160 on the higher-specificity .wx-root, so a
  // plain `page` override loses at runtime and brand/cart-bar/badges leak back to
  // bright green. !important on the custom-property declaration is the fix — guard
  // it so a reformat/lint-autofix that drops it can't silently reintroduce the bug.
  it('pins --weui-BRAND with !important to beat weui\'s .wx-root green', () => {
    expect(css).toMatch(/--weui-BRAND:\s*#FF4896\s*!important/i)
  })

  it('leaves no Pine-Ink residue in the token block', () => {
    for (const stale of ['#234B3A', '#2F6B4F', '#C8643C', '#B0491F']) {
      expect(css).not.toContain(stale)
    }
  })
})

describe('theme tokens — WCAG contrast floors', () => {
  const BG0 = token('--weui-BG-0')
  const BG1 = token('--weui-BG-1')
  const BRAND = token('--weui-BRAND')
  const PINK_DEEP = token('--green-2')
  const FG0 = token('--weui-FG-0')
  const FG2 = token('--weui-FG-2')
  const PRICE = token('--price-ink')
  const TITLE = token('--jf-title-blue')
  const TAG_PINK = token('--jf-tag-pink')
  const TAG_RED = token('--jf-tag-red')
  const WHITE = '#FFFFFF'

  it('price red passes AA on both light backgrounds (>=4.5:1)', () => {
    expect(contrast(PRICE, BG0)).toBeGreaterThanOrEqual(4.5)
    expect(contrast(PRICE, BG1)).toBeGreaterThanOrEqual(4.5)
  })
  it('secondary text passes AA for small text on the page (>=4.5:1)', () => {
    expect(contrast(FG2, BG0)).toBeGreaterThanOrEqual(4.5)
    expect(contrast(FG2, BG1)).toBeGreaterThanOrEqual(4.5)
  })
  it('body text is high-contrast on the page (>=7:1)', () => {
    expect(contrast(FG0, BG0)).toBeGreaterThanOrEqual(7.0)
  })
  it('white small text on the deep-pink band passes AA (>=4.5:1)', () => {
    expect(contrast(WHITE, PINK_DEEP)).toBeGreaterThanOrEqual(4.5)
  })
  it('white large brand text on bright pink passes AA-large (>=3:1)', () => {
    expect(contrast(WHITE, BRAND)).toBeGreaterThanOrEqual(3.0)
  })
  it('# section title (blue) passes AA on the card (>=4.5:1)', () => {
    expect(contrast(TITLE, BG1)).toBeGreaterThanOrEqual(4.5)
  })
  it('coupon text passes AA on the coupon pink (>=4.5:1)', () => {
    expect(contrast(TAG_RED, TAG_PINK)).toBeGreaterThanOrEqual(4.5)
  })
  it('gold-card ink passes AA on the gold gradient deep end (>=4.5:1) — 设计金卡的折中达标', () => {
    expect(contrast(token('--jf-gold-ink'), token('--jf-gold-bg-1'))).toBeGreaterThanOrEqual(4.5)
  })

  // 防回归：黄色药丸（左轨数量徽章/去结算）里的文字必须深色。白字 on 黄 = 1.44:1，
  // 此前白字徽章把真实分类数量渲染成不可读（/ship 拦截）。深蓝字 on 黄 ≈ 11:1。
  const ACCENT = token('--accent')
  it('yellow-pill text (title-blue) passes AA for small text (>=4.5:1)', () => {
    expect(contrast(TITLE, ACCENT)).toBeGreaterThanOrEqual(4.5)
  })
  // 价格数字 on 黄底是大字粗体，只需 AA-large 3:1（既有折中，价签数字够大够粗）
  it('price red on the yellow tag meets AA-large for big bold numerals (>=3:1)', () => {
    expect(contrast(PRICE, ACCENT)).toBeGreaterThanOrEqual(3.0)
  })
})
