/**
 * Delivery API: getDeliveryQuote posts the recipient + coords; createOrder
 * threads delivery info + quote_token only for delivery orders (T7).
 */
global.wx = {
  getAccountInfoSync: () => ({ miniProgram: { envVersion: 'develop' } }),
  getStorageSync: jest.fn(() => 'tok'),
  removeStorageSync: jest.fn(),
  request: jest.fn(),
}

const { getDeliveryQuote, createOrder } = require('../api/index.js')

beforeEach(() => {
  jest.clearAllMocks()
})

describe('getDeliveryQuote', () => {
  it('posts shop_id + recipient + coords to /delivery/quote', () => {
    getDeliveryQuote(1, {
      recipient_name: '张三',
      recipient_phone: '13800000000',
      recipient_address: '北京市朝阳区某路1号',
      lat: 39.9,
      lng: 116.4,
    })
    const call = wx.request.mock.calls[0][0]
    expect(call.url).toContain('/delivery/quote')
    expect(call.method).toBe('POST')
    expect(call.data.shop_id).toBe(1)
    expect(call.data.recipient_lat).toBe(39.9)
    expect(call.data.recipient_lng).toBe(116.4)
    expect(call.data.recipient_name).toBe('张三')
  })
})

describe('createOrder delivery payload', () => {
  it('includes delivery + quote_token for delivery orders', () => {
    const delivery = { recipient_name: '张三', lat: 39.9, lng: 116.4 }
    createOrder(1, '', 100, [], false, 'delivery', delivery, 'TOK123')
    const data = wx.request.mock.calls[0][0].data
    expect(data.order_type).toBe('delivery')
    expect(data.delivery).toEqual(delivery)
    expect(data.quote_token).toBe('TOK123')
  })

  it('omits delivery fields for dine_in orders', () => {
    createOrder(1, 'A01', 100, [], false)
    const data = wx.request.mock.calls[0][0].data
    expect(data.order_type).toBe('dine_in')
    expect(data.delivery).toBeUndefined()
    expect(data.quote_token).toBeUndefined()
  })
})
