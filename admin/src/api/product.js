import client from './client'

// Returns all products across the merchant's shops (all statuses).
export const getMerchantProducts = () => client.get('/merchant/products')

export const createProduct = (data) => client.post('/merchant/products', data)

export const updateProduct = (id, data) => client.put(`/merchant/products/${id}`, data)

export const deleteProduct = (id) => client.delete(`/merchant/products/${id}`)

// Product specs (variants)
export const createProductSpec = (productId, data) =>
  client.post(`/merchant/products/${productId}/specs`, data)

export const updateProductSpec = (id, data) => client.put(`/merchant/specs/${id}`, data)

export const deleteProductSpec = (id) => client.delete(`/merchant/specs/${id}`)
