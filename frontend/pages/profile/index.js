// pages/profile/index.js
const { getBalance, getWalletLogs, getOrders, getInviteStats, getRewardLogs, getRewardExpiryInfo } = require('../../api/index.js')

Page({
  data: {
    user: {},
    balance: { balance: 0, reward_balance: 0 },
    stats: { invite_count: 0, total_invite_reward: 0, today_reward: 0 },
    logs: [],
    orders: [],
    rewardLogs: [],
    rewardPaused: false,
    expiringSoonCount: 0,
    activeTab: 'orders',
    balanceText: '0.00',
    rewardBalanceText: '0.00',
    todayRewardText: '0.00',
    totalInviteRewardText: '0.00',
    tabbar: {
      current: 2,
      list: [
        { text: '点餐', key: 'home' },
        { text: '邀请', key: 'invite' },
        { text: '我的', key: 'profile' }
      ]
    }
  },

  onShow() {
    this.loadData()
  },

  tabChange(e) {
    const index = e.detail.index
    this.setData({ 'tabbar.current': index })
    const routes = ['/pages/home/index', '/pages/invite/index', '/pages/profile/index']
    wx.reLaunch({ url: routes[index] + '?fromTabbar=1' })
  },

  loadData() {
    const user = wx.getStorageSync('user') || {}
    Promise.all([
      getBalance(),
      getInviteStats(),
      getWalletLogs(),
      getOrders(),
      getRewardLogs(),
      getRewardExpiryInfo()
    ]).then(([balance, stats, logsRes, ordersRes, rewardLogsRes, expiryInfo]) => {
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
        return {
          ...order,
          items,
          amountText: '¥' + order.amount.toFixed(2),
          statusText: ['', '待支付', '已完成', '已完成'][order.status] || '已完成',
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
        balance,
        stats,
        logs,
        orders,
        rewardLogs,
        rewardPaused: expiryInfo.reward_paused || false,
        expiringSoonCount: expiryInfo.expiring_soon_count || 0,
        balanceText: (balance.balance || 0).toFixed(2),
        rewardBalanceText: (balance.reward_balance || 0).toFixed(2),
        todayRewardText: (stats.today_reward || 0).toFixed(2),
        totalInviteRewardText: (stats.total_invite_reward || 0).toFixed(2)
      })
    }).catch(err => {
      console.error(err)
    })
  },

  switchTab(e) {
    this.setData({ activeTab: e.currentTarget.dataset.tab })
  },

  goShareCode() {
    wx.navigateTo({ url: '/pages/share-code/index' })
  }
})