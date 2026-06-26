/**
 * Custom SVG tab bar (v6): replaces weui mp-tabbar so the bottom icons match
 * the design comp exactly — line glyphs (店铺/加人/人像), BRAND pine-green when
 * active, muted grey when not. Icons are inline SVG data-URIs (no PNG assets).
 */
const fs = require('fs')
const path = require('path')
const read = (rel) => fs.readFileSync(path.join(__dirname, '..', rel), 'utf8')

const pages = ['home', 'invite', 'menu', 'profile']

describe('custom tabbar replaces mp-tabbar', () => {
  it.each(pages)('%s wraps <tabbar> in .tabbar-fixed and drops mp-tabbar', (p) => {
    const wxml = read(`pages/${p}/index.wxml`)
    expect(wxml).toMatch(/tabbar-fixed[^>]*>\s*<tabbar\b/)
    expect(wxml).not.toMatch(/<mp-tabbar/)
  })

  it('is registered as a global component', () => {
    const appJson = JSON.parse(read('app.json'))
    expect(appJson.usingComponents.tabbar).toBe('components/tabbar/index')
  })
})

describe('tabbar component', () => {
  const wxml = read('components/tabbar/index.wxml')
  const wxss = read('components/tabbar/index.wxss')
  const js = read('components/tabbar/index.js')

  it('renders 点餐 / 邀请 / 我的 tappable items bound to current', () => {
    expect(js).toMatch(/点餐/)
    expect(js).toMatch(/邀请/)
    expect(js).toMatch(/我的/)
    expect(wxml).toMatch(/bindtap="onTap"/)
    expect(wxml).toMatch(/current === index \? 'ctab-t_on'/)
  })

  it('emits a change event carrying the tapped index', () => {
    expect(js).toMatch(/this\.triggerEvent\('change',\s*\{\s*index\s*\}\)/)
  })

  // colour is baked into each SVG data-uri, so guard every icon in both states
  it.each(['menu', 'invite', 'profile'])(
    '%s active icon is BRAND pine (#234B3A), inactive is muted (#8A8275)',
    (icon) => {
      expect(wxss).toMatch(new RegExp(`\\.ctab-t_on[^{]*\\.ctab-ic_${icon}[^}]*%23234B3A`))
      expect(wxss).toMatch(new RegExp(`\\.ctab-ic_${icon}\\s*\\{[^}]*%238A8275`))
    }
  )

  it('uses no PNG assets and no bright weui green (#07c160)', () => {
    expect(wxss).not.toMatch(/\.png/)
    expect(wxss).not.toMatch(/%2307c160/i)
  })
})
