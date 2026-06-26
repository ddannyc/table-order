/**
 * Behavioral contract for the custom <tabbar> (replaces mp-tabbar).
 * The component emits change{index}; every tab page's tabChange must map that
 * bare integer to the same fixed route and highlight the right tab. After the
 * swap this contract was untested — a reordered items[] or routes[] would break
 * navigation silently. Mirrors the page-harness pattern in home-launcher.test.js.
 */
const fs = require('fs')
const path = require('path')

global.wx = {
  getAccountInfoSync: jest.fn(() => ({ miniProgram: { envVersion: 'develop' } })),
  getStorageSync: jest.fn(() => ''),
  setStorageSync: jest.fn(),
  removeStorageSync: jest.fn(),
  reLaunch: jest.fn(),
  navigateTo: jest.fn(),
  switchTab: jest.fn(),
  request: jest.fn(),
  showToast: jest.fn(),
  showModal: jest.fn(),
}
global.getApp = () => ({})

let pageConfig
global.Page = (config) => {
  pageConfig = config
}

const load = (name) => {
  jest.isolateModules(() => {
    require(`../pages/${name}/index.js`)
  })
  return pageConfig
}
const read = (rel) => fs.readFileSync(path.join(__dirname, '..', rel), 'utf8')

const EXPECTED_CURRENT = { home: 0, menu: 0, invite: 1, profile: 2 }
const ROUTES = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
const PAGES = Object.keys(EXPECTED_CURRENT)

beforeEach(() => jest.clearAllMocks())

describe('tabbar routing contract', () => {
  it.each(PAGES)('%s highlights its own tab via tabbarCurrent', (name) => {
    expect(load(name).data.tabbarCurrent).toBe(EXPECTED_CURRENT[name])
  })

  it.each(PAGES)('%s binds current + bindchange on <tabbar>', (name) => {
    const wxml = read(`pages/${name}/index.wxml`)
    expect(wxml).toMatch(/<tabbar[^>]*current="\{\{tabbarCurrent\}\}"/)
    expect(wxml).toMatch(/<tabbar[^>]*bindchange="tabChange"/)
  })

  it.each(PAGES)('%s tabChange routes each tapped index to the shared destination', (name) => {
    const cfg = load(name)
    ;[0, 1, 2].forEach((index) => {
      jest.clearAllMocks()
      const setData = jest.fn()
      cfg.tabChange.call({ setData }, { detail: { index } })
      expect(setData).toHaveBeenCalledWith({ tabbarCurrent: index })
      expect(wx.reLaunch).toHaveBeenCalledWith({ url: ROUTES[index] + '?fromTabbar=1' })
    })
  })
})
