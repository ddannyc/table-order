// pages/profile/index.js
const { getWalletLogs, getOrders, getInviteStats, getRewardBalance, getRewardLogs, getRewardExpiryInfo } = require('../../api/index.js')
const { doLogin, handleAuthError } = require('../../utils/storage.js')
const { TAB_LIST } = require('../../utils/tabbar.js')

Page({
  data: {
    needLogin: false,
    user: {},
    stats: { invite_count: 0, total_invite_reward: 0, today_reward: 0 },
    logs: [],
    orders: [],
    rewardLogs: [],
    rewardPaused: false,
    expiringSoonCount: 0,
    activeTab: 'orders',
    rewardBalanceText: '0.00',
    todayRewardText: '0.00',
    totalInviteRewardText: '0.00',
    tabbarList: TAB_LIST,
    tabbarCurrent: 2
  },

  onShow() {
    const token = wx.getStorageSync('token')
    if (!token) {
      this.setData({ needLogin: true })
      return
    }
    this.setData({ needLogin: false })
    this.loadData()
  },

  handleLogin() {
    doLogin()
      .then(() => {
        this.setData({ needLogin: false })
        this.loadData()
      })
      .catch(() => {})
  },

  tabChange(e) {
    const index = e.detail.index
    this.setData({ tabbarCurrent: index })
    const routes = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
    wx.reLaunch({ url: routes[index] + '?fromTabbar=1' })
  },

  loadData() {
    const user = wx.getStorageSync('user') || {}
    Promise.all([
      getRewardBalance(),
      getInviteStats(),
      getWalletLogs(),
      getOrders(),
      getRewardLogs(),
      getRewardExpiryInfo()
    ]).then(([rewardData, stats, logsRes, ordersRes, rewardLogsRes, expiryInfo]) => {
      const typeLabelMap = { reward: '返利', invite_reward: '邀请奖励', deduct: '抵扣' }
      const logs = (logsRes.logs || logsRes || []).map(log => ({
        ...log,
        amountText: (log.amount < 0 ? '-' : '+') + '¥' + Math.abs(log.amount).toFixed(2),
        typeLabel: typeLabelMap[log.type] || log.type
      }))
      const orders = (ordersRes.orders || ordersRes || []).map(order => {
        const d = new Date(order.created_at)
        const dateText = `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
        const items = (order.items || []).map(item => ({
          ...item,
          subtotalText: '¥' + (item.subtotal || 0).toFixed(2)
        }))
        const isDelivery = order.order_type === 'delivery'
        return {
          ...order,
          items,
          amountText: '¥' + order.amount.toFixed(2),
          statusText: ['', '待支付', '已完成', '已完成'][order.status] || '已完成',
          deliveryText: isDelivery && order.delivery ? '外卖 · ' + (order.delivery.status_label || '配送中') : '',
          dateText
        }
      })
      const rewardLogs = (rewardLogsRes.logs || []).map(log => ({
        ...log,
        amountText: '+' + log.amount.toFixed(2),
        typeLabel: log.type_label || log.type,
        dateText: (log.created_at || '').substring(0, 10)
      }))
      this.setData({
        user,
        stats,
        logs,
        orders,
        rewardLogs,
        rewardPaused: expiryInfo.reward_paused || false,
        expiringSoonCount: expiryInfo.expiring_soon_count || 0,
        rewardBalanceText: (rewardData.reward_balance || 0).toFixed(2),
        todayRewardText: (stats.today_reward || 0).toFixed(2),
        totalInviteRewardText: (stats.total_invite_reward || 0).toFixed(2)
      })
    }).catch(err => {
      if (!handleAuthError(err, this)) { console.error(err) }
    })
  },

  switchTab(e) {
    this.setData({ activeTab: e.currentTarget.dataset.tab })
  },

  goShareCode() {
    wx.navigateTo({ url: '/pages/share-code/index' })
  }
})