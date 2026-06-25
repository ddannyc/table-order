import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('./client', () => ({
  default: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

import client from './client'
import { createProductSpec, updateProductSpec, deleteProductSpec } from './product'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('product spec api', () => {
  it('createProductSpec posts to the product specs endpoint', () => {
    createProductSpec(7, { name: '600ml', price: 15 })
    expect(client.post).toHaveBeenCalledWith('/merchant/products/7/specs', {
      name: '600ml',
      price: 15,
    })
  })

  it('updateProductSpec puts to the spec endpoint', () => {
    updateProductSpec(3, { price: 18 })
    expect(client.put).toHaveBeenCalledWith('/merchant/specs/3', { price: 18 })
  })

  it('deleteProductSpec deletes the spec endpoint', () => {
    deleteProductSpec(3)
    expect(client.delete).toHaveBeenCalledWith('/merchant/specs/3')
  })
})
