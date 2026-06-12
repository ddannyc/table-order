/**
 * Tests for utils/auth.js — isLoggedIn, handleAuthError, doLogin, requireLogin
 */

// Mock wx APIs
global.wx = {
  getStorageSync: jest.fn(),
  setStorageSync: jest.fn(),
  removeStorageSync: jest.fn(),
  login: jest.fn(),
  showToast: jest.fn(),
  navigateTo: jest.fn(),
  setClipboardData: jest.fn(),
}

// Mock API module
jest.mock('../api/index.js', () => ({
  loginByCode: jest.fn(),
  bindInviteCode: jest.fn(() => Promise.resolve()),
}))

// auth.js no longer imports from storage.js (circular dep fix).
// It uses wx.getStorageSync / wx.setStorageSync directly,
// which are already mocked in global.wx above.

const { loginByCode, bindInviteCode } = require('../api/index.js')

describe('isLoggedIn', () => {
  let isLoggedIn

  beforeEach(() => {
    jest.clearAllMocks()
    jest.isolateModules(() => {
      isLoggedIn = require('../utils/auth.js').isLoggedIn
    })
  })

  it('returns true when token exists', () => {
    wx.getStorageSync.mockReturnValue('valid-token')
    expect(isLoggedIn()).toBe(true)
  })

  it('returns false when token is empty string', () => {
    wx.getStorageSync.mockReturnValue('')
    expect(isLoggedIn()).toBe(false)
  })

  it('returns false when token is null/undefined', () => {
    wx.getStorageSync.mockReturnValue(null)
    expect(isLoggedIn()).toBe(false)
  })
})

describe('handleAuthError', () => {
  let handleAuthError

  beforeEach(() => {
    jest.clearAllMocks()
    jest.isolateModules(() => {
      handleAuthError = require('../utils/auth.js').handleAuthError
    })
  })

  it('sets needLogin=true and returns true on 401 error', () => {
    const page = { setData: jest.fn() }
    const err = new Error('未登录')
    const result = handleAuthError(err, page)
    expect(result).toBe(true)
    expect(page.setData).toHaveBeenCalledWith({ needLogin: true })
  })

  it('returns false and does not modify page for non-401 errors', () => {
    const page = { setData: jest.fn() }
    const err = new Error('Network error')
    const result = handleAuthError(err, page)
    expect(result).toBe(false)
    expect(page.setData).not.toHaveBeenCalled()
  })

  it('returns false when err is null/undefined', () => {
    const page = { setData: jest.fn() }
    expect(handleAuthError(null, page)).toBe(false)
    expect(handleAuthError(undefined, page)).toBe(false)
    expect(page.setData).not.toHaveBeenCalled()
  })
})

describe('doLogin', () => {
  let doLogin

  beforeEach(() => {
    jest.clearAllMocks()
    jest.isolateModules(() => {
      doLogin = require('../utils/auth.js').doLogin
    })
  })

  it('calls wx.login, then loginByCode with the code', async () => {
    wx.login.mockImplementation(({ success }) => success({ code: 'test-code' }))
    loginByCode.mockResolvedValue({ token: 'abc', user: {} })

    await doLogin()

    expect(wx.login).toHaveBeenCalled()
    expect(loginByCode).toHaveBeenCalledWith('test-code')
  })

  it('stores token and user on success', async () => {
    wx.login.mockImplementation(({ success }) => success({ code: 'test-code' }))
    loginByCode.mockResolvedValue({ token: 'abc123', user: { name: 'Test' } })

    await doLogin()

    expect(wx.setStorageSync).toHaveBeenCalledWith('token', 'abc123')
    expect(wx.setStorageSync).toHaveBeenCalledWith('user', { name: 'Test' })
  })

  it('binds pending_invite_code if present, then removes it', async () => {
    wx.login.mockImplementation(({ success }) => success({ code: 'test-code' }))
    loginByCode.mockResolvedValue({ token: 'abc', user: {} })
    wx.getStorageSync.mockReturnValue('pending-xxx')

    await doLogin()

    // Should read pending code
    expect(wx.getStorageSync).toHaveBeenCalledWith('pending_invite_code')
    // Should remove before binding
    expect(wx.removeStorageSync).toHaveBeenCalledWith('pending_invite_code')
    // Should call bindInviteCode with the pending code
    expect(bindInviteCode).toHaveBeenCalledWith('pending-xxx')
  })

  it('does not attempt bind when no pending code', async () => {
    wx.login.mockImplementation(({ success }) => success({ code: 'test-code' }))
    loginByCode.mockResolvedValue({ token: 'abc', user: {} })
    wx.getStorageSync.mockReturnValue(null)

    await doLogin()

    expect(bindInviteCode).not.toHaveBeenCalled()
  })

  it('rejects and shows toast when loginByCode fails', async () => {
    wx.login.mockImplementation(({ success }) => success({ code: 'test-code' }))
    loginByCode.mockRejectedValue(new Error('Server error'))

    await expect(doLogin()).rejects.toThrow('Server error')
    expect(wx.showToast).toHaveBeenCalledWith({ title: '登录失败', icon: 'none' })
  })

  it('rejects and shows toast when response has no token', async () => {
    wx.login.mockImplementation(({ success }) => success({ code: 'test-code' }))
    // Backend returns 200 but no token field
    loginByCode.mockResolvedValue({ user: { name: 'Test' } })

    await expect(doLogin()).rejects.toThrow('登录失败：token 缺失')
    expect(wx.showToast).toHaveBeenCalledWith({ title: '登录失败', icon: 'none' })
  })

  it('rejects and shows toast when wx.login fails', async () => {
    wx.login.mockImplementation(({ fail }) => fail())

    await expect(doLogin()).rejects.toThrow('wx.login failed')
    expect(wx.showToast).toHaveBeenCalledWith({ title: '微信登录失败', icon: 'none' })
  })
})

describe('requireLogin', () => {
  let requireLogin

  beforeEach(() => {
    jest.clearAllMocks()
    jest.isolateModules(() => {
      requireLogin = require('../utils/auth.js').requireLogin
    })
  })

  it('returns true when token exists', () => {
    wx.getStorageSync.mockReturnValue('valid-token')
    const result = requireLogin()
    expect(result).toBe(true)
    expect(wx.navigateTo).not.toHaveBeenCalled()
  })

  it('returns false and navigates to login when no token', () => {
    wx.getStorageSync.mockReturnValue('')
    global.getCurrentPages = jest.fn(() => [])
    const result = requireLogin()
    expect(result).toBe(false)
    expect(wx.navigateTo).toHaveBeenCalledWith({ url: '/pages/login/index' })
  })

  it('stores return_path when no token', () => {
    wx.getStorageSync.mockReturnValue('')
    global.getCurrentPages = jest.fn(() => [
      { route: 'pages/profile/index', options: {} }
    ])

    requireLogin()

    expect(wx.setStorageSync).toHaveBeenCalledWith('return_path', '/pages/profile/index')
  })
})
