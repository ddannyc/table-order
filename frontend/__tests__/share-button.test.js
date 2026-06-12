/**
 * Tests for "分享到微信" button visibility
 *
 * RED: Tests fail — no Page() mock or setData binding
 * GREEN: Tests pass once Page() mock captures data properly
 *
 * Verify: share buttons hidden when needLogin=true, visible when needLogin=false
 */

// Mock wx APIs used by the pages
global.wx = {
  getStorageSync: jest.fn(),
  setStorageSync: jest.fn(),
  removeStorageSync: jest.fn(),
  login: jest.fn(),
  request: jest.fn(),
  showToast: jest.fn(),
  arrayBufferToBase64: jest.fn(() => ''),
  navigateTo: jest.fn(),
  reLaunch: jest.fn(),
  setClipboardData: jest.fn(),
}

// Mock dependencies
jest.mock('../api/index.js', () => ({
  getInviteStats: jest.fn(() => Promise.resolve({ invite_count: 0, total_invite_reward: 0, today_reward: 0 })),
  bindInviteCode: jest.fn(() => Promise.resolve()),
  getInviteQR: jest.fn(() => Promise.resolve(new ArrayBuffer(0))),
  getRewardBalance: jest.fn(() => Promise.resolve({ reward_balance: 0, reward_paused: false })),
}))

jest.mock('../utils/storage.js', () => ({
  doLogin: jest.fn(() => Promise.resolve()),
  handleAuthError: jest.fn(() => false),
}))

/**
 * Helper: create a WeChat Page() mock that captures the config
 * and binds setData to update the data object in-place.
 * Returns { config, data } where config is the Page config object
 * with bound methods, and data is the observable data state.
 */
function createPageMock() {
  const data = {}
  const config = {}

  global.Page = (cfg) => {
    Object.assign(config, cfg)
    config.data = { ...cfg.data }
    // Bind setData to update data object
    config.setData = function (obj) {
      Object.assign(config.data, obj)
    }
    // Bind lifecycle methods so `this` refers to config
    if (config.onLoad) config.onLoad = config.onLoad.bind(config)
    if (config.onShow) config.onShow = config.onShow.bind(config)
    if (config.onShareAppMessage) config.onShareAppMessage = config.onShareAppMessage.bind(config)
    if (config.loadData) config.loadData = config.loadData.bind(config)
  }

  return config
}

describe('invite page — 分享到微信 button', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('sets needLogin=true when no token on load', () => {
    wx.getStorageSync
      .mockReturnValueOnce('') // token is empty (first call)
      .mockReturnValue('')     // all subsequent calls

    const config = createPageMock()

    jest.isolateModules(() => {
      require('../pages/invite/index')
    })

    config.onLoad({})

    expect(config.data.needLogin).toBe(true)
  })

  it('sets needLogin=false when token exists on load', () => {
    wx.getStorageSync
      .mockReturnValueOnce('valid-token') // token exists
      .mockReturnValue({})                 // everything else

    const config = createPageMock()

    jest.isolateModules(() => {
      require('../pages/invite/index')
    })

    config.onLoad({})

    expect(config.data.needLogin).toBe(false)
  })
})

describe('share-code page — 分享给好友 button', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('sets needLogin=true when no token on load', () => {
    wx.getStorageSync
      .mockReturnValueOnce('')  // no token
      .mockReturnValue('')

    const config = createPageMock()

    jest.isolateModules(() => {
      require('../pages/share-code/index')
    })

    config.onLoad()

    expect(config.data.needLogin).toBe(true)
  })

  it('sets needLogin=false when token exists on load', () => {
    wx.getStorageSync
      .mockReturnValueOnce('valid-token') // token exists
      .mockReturnValue({})

    const config = createPageMock()

    jest.isolateModules(() => {
      require('../pages/share-code/index')
    })

    config.onLoad()

    expect(config.data.needLogin).toBe(false)
  })
})
