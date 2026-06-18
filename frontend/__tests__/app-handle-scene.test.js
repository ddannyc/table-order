/**
 * Tests for app.js handleScene — QR code scan routing
 *
 * Three entry paths:
 *   1. URL Scheme direct query params (shop_id + table_no)
 *   2. QR code scene string "shop_id=1&table_no=A01&token=xxx"
 *   3. Invite code scene string "ic=ABC123"
 */

// Build the mock before app.js loads
global.wx = {
  getStorageSync: jest.fn(),
  setStorageSync: jest.fn(),
  removeStorageSync: jest.fn(),
  reLaunch: jest.fn(),
}

// App() is a WeChat global that registers the app. Capture the config object.
let appConfig
global.App = (config) => {
  appConfig = config
}

require('../app.js')

beforeEach(() => {
  jest.clearAllMocks()
})

describe('handleScene — URL Scheme direct query params', () => {
  it('navigates to home page when shop_id and table_no are in query', () => {
    const options = {
      query: { shop_id: '1', table_no: 'A01' },
    }

    // trigger onLaunch (which calls handleScene internally)
    appConfig.onLaunch(options)

    expect(wx.setStorageSync).toHaveBeenCalledWith('current_shop_id', 1)
    expect(wx.setStorageSync).toHaveBeenCalledWith('current_table_no', 'A01')
    expect(wx.removeStorageSync).toHaveBeenCalledWith('pending_invite_code')
    expect(wx.reLaunch).toHaveBeenCalledWith({
      url: '/pages/home/index?shop_id=1&table_no=A01',
    })
  })

  it('does nothing when query has only shop_id', () => {
    appConfig.onLaunch({ query: { shop_id: '1' } })
    expect(wx.reLaunch).not.toHaveBeenCalled()
  })
})

describe('handleScene — QR code scene string (WeChat link rule)', () => {
  it('extracts shop_id and table_no from scene and navigates', () => {
    const options = {
      query: {
        scene: 'shop_id=1&table_no=A01&token=6f28f1a172ed4971e883889590c99c20',
      },
    }

    appConfig.onLaunch(options)

    expect(wx.setStorageSync).toHaveBeenCalledWith('current_shop_id', 1)
    expect(wx.setStorageSync).toHaveBeenCalledWith('current_table_no', 'A01')
    expect(wx.removeStorageSync).toHaveBeenCalledWith('pending_invite_code')
    expect(wx.reLaunch).toHaveBeenCalledWith({
      url: '/pages/home/index?shop_id=1&table_no=A01',
    })
  })

  it('works when params are in different order', () => {
    const options = {
      query: {
        scene: 'token=abc&table_no=B07&shop_id=42',
      },
    }

    appConfig.onShow(options)

    expect(wx.setStorageSync).toHaveBeenCalledWith('current_shop_id', 42)
    expect(wx.setStorageSync).toHaveBeenCalledWith('current_table_no', 'B07')
    expect(wx.reLaunch).toHaveBeenCalledWith({
      url: '/pages/home/index?shop_id=42&table_no=B07',
    })
  })

  it('does nothing when scene has no shop_id', () => {
    appConfig.onLaunch({
      query: { scene: 'table_no=A01&token=abc' },
    })

    expect(wx.reLaunch).not.toHaveBeenCalled()
    expect(wx.setStorageSync).not.toHaveBeenCalledWith('current_shop_id', expect.anything())
  })
})

describe('handleScene — "扫普通链接二维码打开小程序" q param (URL-encoded)', () => {
  it('extracts shop_id and table_no from the encoded scan URL', () => {
    const url = 'https://example.com/scan?shop_id=1&table_no=A01&token=6f28f1a172ed4971e883889590c99c20'
    appConfig.onLaunch({ query: { q: encodeURIComponent(url) } })

    expect(wx.setStorageSync).toHaveBeenCalledWith('current_shop_id', 1)
    expect(wx.setStorageSync).toHaveBeenCalledWith('current_table_no', 'A01')
    expect(wx.removeStorageSync).toHaveBeenCalledWith('pending_invite_code')
    expect(wx.reLaunch).toHaveBeenCalledWith({
      url: '/pages/home/index?shop_id=1&table_no=A01',
    })
  })

  it('handles q on warm start via onShow', () => {
    const url = 'https://example.com/scan?shop_id=42&table_no=B07&token=abc'
    appConfig.onShow({ query: { q: encodeURIComponent(url) } })

    expect(wx.setStorageSync).toHaveBeenCalledWith('current_shop_id', 42)
    expect(wx.setStorageSync).toHaveBeenCalledWith('current_table_no', 'B07')
    expect(wx.reLaunch).toHaveBeenCalledWith({
      url: '/pages/home/index?shop_id=42&table_no=B07',
    })
  })

  it('does nothing when q URL lacks shop_id', () => {
    const url = 'https://example.com/scan?table_no=A01&token=abc'
    appConfig.onLaunch({ query: { q: encodeURIComponent(url) } })
    expect(wx.reLaunch).not.toHaveBeenCalled()
  })
})

describe('handleScene — invite code scene string', () => {
  it('navigates to invite page for ic=CODE format', () => {
    appConfig.onLaunch({
      query: { scene: 'ic=ABC123XYZ' },
    })

    expect(wx.setStorageSync).toHaveBeenCalledWith('pending_invite_code', 'ABC123XYZ')
    expect(wx.reLaunch).toHaveBeenCalledWith({ url: '/pages/invite/index' })
  })
})

describe('handleScene — edge cases', () => {
  it('does nothing when options is null', () => {
    appConfig.onLaunch(null)
    expect(wx.reLaunch).not.toHaveBeenCalled()
  })

  it('does nothing when options.query is missing', () => {
    appConfig.onLaunch({})
    expect(wx.reLaunch).not.toHaveBeenCalled()
  })

  it('does nothing when scene is empty string', () => {
    appConfig.onLaunch({ query: { scene: '' } })
    expect(wx.reLaunch).not.toHaveBeenCalled()
  })

  it('handles warm start via onShow', () => {
    appConfig.onShow({
      query: { scene: 'shop_id=99&table_no=Z99' },
    })

    expect(wx.setStorageSync).toHaveBeenCalledWith('current_shop_id', 99)
    expect(wx.setStorageSync).toHaveBeenCalledWith('current_table_no', 'Z99')
    expect(wx.reLaunch).toHaveBeenCalledWith({
      url: '/pages/home/index?shop_id=99&table_no=Z99',
    })
  })

  it('table QR rule wins over invite when scene contains both patterns', () => {
    appConfig.onLaunch({
      query: { scene: 'ic=shop_id=1&table_no=A01' },
    })

    // Contains both "ic=" and "shop_id=" — table QR rule is checked first and wins
    expect(wx.setStorageSync).toHaveBeenCalledWith('current_shop_id', 1)
    expect(wx.reLaunch).toHaveBeenCalledWith({
      url: '/pages/home/index?shop_id=1&table_no=A01',
    })
  })
})
