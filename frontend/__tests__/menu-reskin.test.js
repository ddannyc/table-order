/**
 * Structural tests for the 松墨 Pine-Ink menu reskin (T4/T5/T6).
 * Reads menu wxml/wxss as text. Behavior/cart isolation are covered by
 * menu-page.test.js and cart-isolation.test.js — this file only guards the skin.
 */
const fs = require('fs')
const path = require('path')

const wxss = fs.readFileSync(path.join(__dirname, '../pages/menu/index.wxss'), 'utf8')
const wxml = fs.readFileSync(path.join(__dirname, '../pages/menu/index.wxml'), 'utf8')

describe('menu reskin — Pine-Ink surfaces (T4)', () => {
  it('price uses the bronze price-ink token', () => {
    expect(wxss).toMatch(/\.menu-price\s*\{[^}]*color:\s*var\(--price-ink\)/)
  })

  it('the checkout bar uses the deep-green brand background', () => {
    expect(wxss).toMatch(/\.menu-cartbar\s*\{[^}]*background:\s*var\(--weui-BRAND\)/)
  })

  it('the content list reads as a near-white panel against the cream page', () => {
    expect(wxss).toMatch(/\.menu-list\s*\{[^}]*background:\s*var\(--weui-BG-1\)/)
  })
})

describe('menu reskin — category line glyphs (T5)', () => {
  it.each(['cup', 'bubble', 'cheese', 'sparkle'])(
    'defines a gold line-glyph for %s',
    (g) => {
      const re = new RegExp('\\.menu-thumb-ph_' + g + '\\s*\\{[^}]*data:image\\/svg\\+xml')
      expect(wxss).toMatch(re)
    }
  )

  it('keeps the text label in the placeholder for recognition/a11y', () => {
    expect(wxml).toMatch(/\{\{p\.ph\.label\}\}/)
  })
})
