import client from './client'

export const getShops = () => client.get('/merchant/shops')

export const createShop = (data) => client.post('/merchant/shops', data)

export const updateShop = (id, data) => client.put(`/merchant/shops/${id}`, data)
