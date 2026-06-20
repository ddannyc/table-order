// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../router', () => ({ default: { push: vi.fn() } }))
vi.mock('element-plus', () => ({ ElMessage: { error: vi.fn() } }))

import client from './client'
import router from '../router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const requestFulfilled = () => client.interceptors.request.handlers[0].fulfilled
const responseRejected = () => client.interceptors.response.handlers[0].rejected

describe('axios client interceptors', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('injects Authorization header when a token is present', () => {
    useAuthStore().setAuth('tok123', { id: 1 })
    const cfg = requestFulfilled()({ headers: {} })
    expect(cfg.headers.Authorization).toBe('Bearer tok123')
  })

  it('does not set Authorization when there is no token', () => {
    const cfg = requestFulfilled()({ headers: {} })
    expect(cfg.headers.Authorization).toBeUndefined()
  })

  it('on 401: logs out and redirects to /login', async () => {
    const auth = useAuthStore()
    auth.setAuth('tok', { id: 1 })
    await expect(responseRejected()({ response: { status: 401 } })).rejects.toBeTruthy()
    expect(auth.token).toBe('')
    expect(router.push).toHaveBeenCalledWith('/login')
  })

  it('on non-401: surfaces the server error and keeps the session', async () => {
    const auth = useAuthStore()
    auth.setAuth('tok', { id: 1 })
    await expect(
      responseRejected()({ response: { status: 500, data: { error: 'boom' } } }),
    ).rejects.toBeTruthy()
    expect(ElMessage.error).toHaveBeenCalledWith('boom')
    expect(auth.token).toBe('tok')
    expect(router.push).not.toHaveBeenCalled()
  })
})
