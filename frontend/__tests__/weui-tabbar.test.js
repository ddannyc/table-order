/**
 * Verifies the 4 tab pages use the weui mp-tabbar component with a shared
 * icon list, replacing the custom tab bar (Task 2).
 */
const fs = require('fs')
const path = require('path')

const read = (rel) => fs.readFileSync(path.join(__dirname, '..', rel), 'utf8')
const pages = ['home', 'invite', 'menu', 'profile']

describe('weui tabbar replaces custom tab bar', () => {
  it.each(pages)('%s uses mp-tabbar, not custom-tab-bar', (p) => {
    const wxml = read(`pages/${p}/index.wxml`)
    expect(wxml).toMatch(/<mp-tabbar/)
    expect(wxml).not.toMatch(/<custom-tab-bar/)
  })
})

describe('shared tab list', () => {
  const { TAB_LIST } = require('../utils/tabbar.js')
  it('has 点餐 / 邀请 / 我的 with icon paths', () => {
    expect(TAB_LIST).toHaveLength(3)
    expect(TAB_LIST.map((t) => t.text)).toEqual(['点餐', '邀请', '我的'])
    TAB_LIST.forEach((t) => {
      expect(t.iconPath).toMatch(/^\/static\/.+\.png$/)
      expect(t.selectedIconPath).toMatch(/^\/static\/.+\.png$/)
    })
  })
})
