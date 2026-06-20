import { defineStore } from 'pinia'

const TOKEN_KEY = 'admin_token'
const MERCHANT_KEY = 'admin_merchant'
const SHOP_KEY = 'admin_shop_id'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    merchant: JSON.parse(localStorage.getItem(MERCHANT_KEY) || 'null'),
    currentShopId: Number(localStorage.getItem(SHOP_KEY)) || null,
  }),
  getters: {
    isAuthenticated: (s) => !!s.token,
  },
  actions: {
    setAuth(token, merchant) {
      this.token = token
      this.merchant = merchant
      localStorage.setItem(TOKEN_KEY, token)
      localStorage.setItem(MERCHANT_KEY, JSON.stringify(merchant))
    },
    setShop(id) {
      this.currentShopId = id
      localStorage.setItem(SHOP_KEY, String(id))
    },
    logout() {
      this.token = ''
      this.merchant = null
      this.currentShopId = null
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(MERCHANT_KEY)
      localStorage.removeItem(SHOP_KEY)
    },
  },
})
