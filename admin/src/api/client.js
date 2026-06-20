import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import router from '../router'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE,
  timeout: 15000,
})

client.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

// Unwrap to response.data on success; surface errors via ElMessage and
// auto-logout on 401.
client.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const status = err.response?.status
    if (status === 401) {
      useAuthStore().logout()
      router.push('/login')
      ElMessage.error('登录已过期，请重新登录')
    } else {
      const msg = err.response?.data?.error || err.message || '请求失败'
      ElMessage.error(msg)
    }
    return Promise.reject(err)
  },
)

export default client
