import client from './client'

// Order action board. params: { shop_id, date, type, status, page, page_size }
export const getMerchantOrders = (params) => client.get('/merchant/orders', { params })

// 出餐: mark a dine-in order's food ready.
export const prepareOrder = (id) => client.post(`/merchant/orders/${id}/prepare`)

// 改状态: manually set order status (1=pending,2=paid,3=completed,4=cancelled).
export const updateOrderStatus = (id, status) =>
  client.put(`/merchant/orders/${id}/status`, { status })

// 重新派单: re-quote + re-dispatch a failed/cancelled delivery order.
export const redispatchOrder = (id) => client.post(`/merchant/orders/${id}/redispatch`)
