/**
 * Tests for 401 recovery — pages set needLogin=true when API returns 401.
 *
 * Each test: mock API to reject with Error('未登录'), load the page,
 * trigger its data-loading path, and verify needLogin was set to true.
 */

// Mock wx APIs
global.wx = {
  getStorageSync: jest.fn(),
  setStorageSync: jest.fn(),
  removeStorageSync: jest.fn(),
  login: jest.fn(),
  request: jest.fn(),
  showToast: jest.fn(),
  showModal: jest.fn(),
  arrayBufferToBase64: jest.fn(() => ''),
  navigateTo: jest.fn(),
  reLaunch: jest.fn(),
  switchTab: jest.fn(),
  setClipboardData: jest.fn(),
  scanCode: jest.fn(),
  requestPayment: jest.fn(),
}

// Mock product module (order-confirm imports getCart, getCartTotal, clearCart)
jest.mock('../api/product.js', () => ({
  getCart: jest.fn(() => []),
  getCartTotal: jest.fn(() => 0),
  clearCart: jest.fn(),
}))

// Provide real handleAuthError logic inline to avoid circular mock dependencies
const handleAuthError = (err, page) => {
  if (err && err.message === '未登录') {
    page.setData({ needLogin: true })
    return true
  }
  return false
}

jest.mock('../utils/storage.js', () => ({
  doLogin: jest.fn(() => Promise.resolve()),
  handleAuthError,
}))

/**
 * Helper: create a WeChat Page() mock that captures the config
 * and binds setData to update the data object in-place.
 */
function createPageMock() {
  const config = {}

  global.Page = (cfg) => {
    Object.assign(config, cfg)
    config.data = { ...cfg.data }
    config.setData = function (obj) {
      Object.assign(config.data, obj)
    }
    // Bind lifecycle methods so `this` refers to config
    if (config.onLoad) config.onLoad = config.onLoad.bind(config)
    if (config.onShow) config.onShow = config.onShow.bind(config)
    if (config.onShareAppMessage) config.onShareAppMessage = config.onShareAppMessage.bind(config)
    if (config.loadData) config.loadData = config.loadData.bind(config)
    if (config.handleLogin) config.handleLogin = config.handleLogin.bind(config)
    if (config.checkAuthAndLoad) config.checkAuthAndLoad = config.checkAuthAndLoad.bind(config)
    if (config.refreshCartDisplay) config.refreshCartDisplay = config.refreshCartDisplay.bind(config)
  }

  return config
}

describe('invite page — 401 recovery', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('sets needLogin=true when getInviteStats returns 401', () => {
    // getInviteStats rejects with 401 (others never called since loadData aborts on error)
    jest.doMock('../api/index.js', () => ({
      getInviteStats: jest.fn(() => Promise.reject(new Error('未登录'))),
      bindInviteCode: jest.fn(() => Promise.resolve()),
      getInviteQR: jest.fn(() => Promise.reject(new Error('未登录'))),
      getRewardBalance: jest.fn(() => Promise.reject(new Error('未登录'))),
    }))

    wx.getStorageSync.mockReturnValue('valid-token')

    const config = createPageMock()
    jest.isolateModules(() => {
      require('../pages/invite/index')
    })

    // Trigger the 401: onLoad passes token check, calls loadData, getInviteStats rejects
    return new Promise(resolve => {
      // Wait for the microtask — getInviteStats is fire-and-forget (no await)
      config.onLoad({})
      // loadData fires getInviteStats(); its catch runs asynchronously
      setTimeout(() => {
        expect(config.data.needLogin).toBe(true)
        resolve()
      }, 50)
    })
  })
})

describe('profile page — 401 recovery', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('sets needLogin=true when Promise.all rejects with 401', () => {
    jest.doMock('../api/index.js', () => ({
      getWalletLogs: jest.fn(() => Promise.reject(new Error('未登录'))),
      getOrders: jest.fn(() => Promise.resolve([])),
      getInviteStats: jest.fn(() => Promise.resolve({})),
      getRewardBalance: jest.fn(() => Promise.resolve({})),
      getRewardLogs: jest.fn(() => Promise.resolve([])),
      getRewardExpiryInfo: jest.fn(() => Promise.resolve({})),
    }))

    wx.getStorageSync.mockReturnValue('valid-token')

    const config = createPageMock()
    jest.isolateModules(() => {
      require('../pages/profile/index')
    })

    // onShow checks token, passes, calls loadData
    return new Promise(resolve => {
      config.onShow()
      setTimeout(() => {
        expect(config.data.needLogin).toBe(true)
        resolve()
      }, 50)
    })
  })
})

describe('order-confirm page — 401 recovery', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('sets needLogin=true when loadData returns 401', () => {
    jest.doMock('../api/index.js', () => ({
      getShop: jest.fn(() => Promise.reject(new Error('未登录'))),
      getRewardBalance: jest.fn(() => Promise.resolve({})),
      getTableBinding: jest.fn(() => ({ shopId: 1, tableNo: 'A01' })),
      createOrder: jest.fn(),
    }))

    wx.getStorageSync.mockReturnValue('valid-token')

    const config = createPageMock()

    // Avoid the "请先扫码绑定桌号" modal flow — set shopId via options
    jest.isolateModules(() => {
      require('../pages/order-confirm/index')
    })

    return new Promise(resolve => {
      config.onLoad({ shop_id: '1', table_no: 'A01' })
      setTimeout(() => {
        expect(config.data.needLogin).toBe(true)
        resolve()
      }, 50)
    })
  })
})

describe('share-code page — 401 recovery', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('sets needLogin=true when getInviteQR returns 401', () => {
    jest.doMock('../api/index.js', () => ({
      getRewardBalance: jest.fn(() => Promise.reject(new Error('未登录'))),
      getInviteQR: jest.fn(() => Promise.reject(new Error('未登录'))),
    }))

    wx.getStorageSync.mockReturnValue('valid-token')

    const config = createPageMock()
    jest.isolateModules(() => {
      require('../pages/share-code/index')
    })

    return new Promise(resolve => {
      config.onLoad()
      setTimeout(() => {
        expect(config.data.needLogin).toBe(true)
        resolve()
      }, 50)
    })
  })
})
