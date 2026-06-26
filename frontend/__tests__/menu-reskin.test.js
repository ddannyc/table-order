/**
 * Structural tests for the 松墨 Pine-Ink menu reskin (T4/T5/T6).
 * Reads menu wxml/wxss as text. Behavior/cart isolation are covered by
 * menu-page.test.js and cart-isolation.test.js — this file only guards the skin.
 */
const fs = require('fs')
const path = require('path')

const wxss = fs.readFileSync(path.join(__dirname, '../pages/menu/index.wxss'), 'utf8')
const wxml = fs.readFileSync(path.join(__dirname, '../pages/menu/index.wxml'), 'utf8')
const art = fs.readFileSync(path.join(__dirname, '../pages/menu/menu-art.wxss'), 'utf8')

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

describe('menu — category placeholder illustrations (D3, in menu-art.wxss)', () => {
  it.each(['cup', 'bubble', 'cheese', 'sparkle'])(
    'defines a colored drink illustration for %s',
    (g) => {
      const re = new RegExp('\\.menu-thumb-ph_' + g + '\\s*\\{[^}]*data:image\\/svg\\+xml')
      expect(art).toMatch(re)
    }
  )

  it('keeps the text label in the placeholder for recognition/a11y', () => {
    expect(wxml).toMatch(/\{\{p\.ph\.label\}\}/)
  })
})

describe('menu reskin — unbound empty-state illustration (T6)', () => {
  it('shows a gold line illustration in the unbound-table state', () => {
    expect(wxml).toMatch(/menu-empty-illu/)
    expect(wxss).toMatch(/\.menu-empty-illu\s*\{[^}]*data:image\/svg\+xml/)
  })

  it('keeps the call-to-action active and specific', () => {
    expect(wxml).toMatch(/扫.*点餐|点餐/)
  })
})

describe('menu reskin — photo-first cards + category counts (R3)', () => {
  it('left rail shows a per-category count badge', () => {
    expect(wxml).toMatch(/categoryCounts/)
    expect(wxss).toMatch(/\.menu-rail-count/)
  })

  it('the no-spec add control is a round button, not a rectangular weui mini button', () => {
    expect(wxss).toMatch(/\.menu-add-round/)
    expect(wxml).toMatch(/menu-add-round/)
  })

  it('placeholder illustrations are colored (brand green outline), not the old gold single-line', () => {
    expect(art).toMatch(/%23234B3A/i) // brand-green outline in the drink illustrations
    expect(art).not.toMatch(/%23C98A2B/i) // no leftover gold
    expect(wxss).not.toMatch(/%23C98A2B/i)
  })
})
