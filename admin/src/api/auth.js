import client from './client'

export function merchantLogin(phone, password) {
  return client.post('/merchant/login', { phone, password })
}
