import client from './client'

// Merchant-wide cumulative totals: { shops, total_users, total_orders, total_revenue }
export const getDashboard = () => client.get('/merchant/dashboard')

// Single-day stats: params { date: 'YYYY-MM-DD', shop_id? } -> { new_users, orders, revenue, rewarded }
export const getStats = (params) => client.get('/merchant/stats', { params })
