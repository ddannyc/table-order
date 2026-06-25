/**
 * Verifies weui-miniprogram components are globally registered (Task 1),
 * so pages can use weui-native components for the de-customization migration.
 */
const fs = require('fs')
const path = require('path')

const appJson = JSON.parse(fs.readFileSync(path.join(__dirname, '..', 'app.json'), 'utf8'))

describe('app.json registers weui components', () => {
  const uc = appJson.usingComponents || {}
  it.each([
    ['mp-tabbar', 'miniprogram_npm/weui-miniprogram/tabbar/tabbar'],
    ['mp-searchbar', 'miniprogram_npm/weui-miniprogram/searchbar/searchbar'],
    ['mp-half-screen-dialog', 'miniprogram_npm/weui-miniprogram/half-screen-dialog/half-screen-dialog'],
    ['mp-dialog', 'miniprogram_npm/weui-miniprogram/dialog/dialog'],
    ['mp-grids', 'miniprogram_npm/weui-miniprogram/grids/grids'],
  ])('registers %s', (name, p) => {
    expect(uc[name]).toBe(p)
  })
})
