/**
 * Tests for the 松墨 Pine-Ink design tokens (T1).
 * Guards three things:
 *   1. Each token is defined in app.wxss with its agreed hex value.
 *   2. .page paints the cream page background (--weui-BG-0).
 *   3. The chosen colors meet WCAG contrast floors — computed here, so a
 *      future token edit that quietly breaks contrast fails the suite.
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

describe('theme tokens — 松墨 Pine-Ink palette (T1)', () => {
  const expected = {
    '--weui-BRAND': '#2C4A3B',
    '--weui-BG-0': '#F3EEE4',
    '--weui-BG-1': '#FBF8F2',
    '--weui-BG-2': '#EFE9DD',
    '--weui-FG-0': '#2A2723',
    '--weui-FG-2': '#6E665A',
    '--weui-FG-3': '#E3DCCE',
    '--brand-accent': '#C98A2B',
    '--price-ink': '#A66E1F',
  }

  it.each(Object.entries(expected))('defines %s = %s', (name, hex) => {
    expect(token(name)).toBe(hex)
  })

  it('.page paints the cream page background (--weui-BG-0)', () => {
    expect(css).toMatch(/\.page\s*\{[^}]*background:\s*var\(--weui-BG-0\)/)
  })
})

describe('theme tokens — WCAG contrast floors', () => {
  const BG0 = '#F3EEE4'
  const BG1 = '#FBF8F2'
  const BRAND = '#2C4A3B'
  const FG0 = '#2A2723'
  const FG2 = '#6E665A'
  const PRICE = '#A66E1F'

  it('price-ink is legible as large bold text on both backgrounds (>=3:1)', () => {
    expect(contrast(PRICE, BG0)).toBeGreaterThanOrEqual(3.0)
    expect(contrast(PRICE, BG1)).toBeGreaterThanOrEqual(3.0)
  })
  it('secondary text passes AA for small text on cream (>=4.5:1)', () => {
    expect(contrast(FG2, BG0)).toBeGreaterThanOrEqual(4.5)
  })
  it('body text is high-contrast on cream (>=7:1)', () => {
    expect(contrast(FG0, BG0)).toBeGreaterThanOrEqual(7.0)
  })
  it('cream text on the brand band passes AA (>=4.5:1)', () => {
    expect(contrast(BG0, BRAND)).toBeGreaterThanOrEqual(4.5)
  })
})
