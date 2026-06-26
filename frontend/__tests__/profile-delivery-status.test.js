/**
 * Profile order list shows the Shansong delivery status for delivery orders (T8).
 */
global.wx = {
  getStorageSync: jest.fn(() => ({})),
  setStorageSync: jest.fn(),
  reLaunch: jest.fn(),
}

jest.mock('../api/index.js', () => ({
  getRewardBalance: jest.fn(() => Promise.resolve({ reward_balance: 0 })),
  getInviteStats: jest.fn(() => Promise.resolve({ invite_count: 0, total_invite_reward: 0, today_reward: 0 })),
  getWalletLogs: jest.fn(() => Promise.resolve({ logs: [] })),
  getRewardLogs: jest.fn(() => Promise.resolve({ logs: [] })),
  getRewardExpiryInfo: jest.fn(() => Promise.resolve({ reward_paused: false, expiring_soon_count: 0 })),
  getOrders: jest.fn(() => Promise.resolve({
    orders: [
      {
        order_no: 'D1', order_type: 'delivery', amount: 108.5, status: 2,
        created_at: '2026-06-26T10:00:00Z', items: [],
        delivery: { status_label: '配送中', delivery_fee: 8.5 },
      },
      {
        order_no: 'I1', order_type: 'dine_in', amount: 50, status: 2,
        created_at: '2026-06-26T10:00:00Z', items: [],
      },
    ],
  })),
}))
jest.mock('../utils/storage.js', () => ({
  doLogin: jest.fn(() => Promise.resolve()),
  handleAuthError: jest.fn(() => false),
}))

let pageConfig
global.Page = (config) => { pageConfig = config }
require('../pages/profile/index.js')

const flush = () => new Promise((r) => setImmediate(r))

it('labels delivery orders with their shansong status and leaves dine-in blank', async () => {
  const ctx = { setData(p) { Object.assign(this.data, p) }, data: {} }
  pageConfig.loadData.call(ctx)
  await flush()
  await flush()

  const orders = ctx.data.orders
  expect(orders).toHaveLength(2)
  expect(orders[0].deliveryText).toBe('外卖 · 配送中')
  expect(orders[1].deliveryText).toBe('')
})
