import client from './client'

// Order action board. params: { shop_id, date, type, status, page, page_size }
export const getMerchantOrders = (params) => client.get('/merchant/orders', { params })
