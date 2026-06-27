/**
 * Structural tests for the order-confirm replication (R4).
 * Visual-only: photo thumbnails in 商品明细 + terracotta line prices.
 * Payment logic (WeChat Pay + reward deduction) is unchanged and covered by
 * order-confirm-type / -delivery / api-create-order tests.
 *
 * Note: a 余额支付 method is intentionally NOT added — the backend order path
 * has no wallet-balance payment, so it would be fake UI ("只复刻有真数据的").
 */
const fs = require('fs')
const path = require('path')

const wxml = fs.readFileSync(path.join(__dirname, '../pages/order-confirm/index.wxml'), 'utf8')
const wxss = fs.readFileSync(path.join(__dirname, '../pages/order-confirm/index.wxss'), 'utf8')

describe('order-confirm reskin — photo thumbnails + terracotta prices (R4)', () => {
  it('renders a thumbnail bound to the cart item image for each line', () => {
    expect(wxml).toMatch(/oc-item-thumb/)
    expect(wxml).toMatch(/item\.image/)
  })

  it('line prices use the terracotta price ink', () => {
    expect(wxss).toMatch(/\.oc-item-price\s*\{[^}]*color:\s*var\(--price-ink\)/)
  })
})

describe('order-confirm — payment method + reward toggle (D4, v6)', () => {
  it('shows WeChat Pay as the only payment method', () => {
    expect(wxml).toMatch(/支付方式/)
    expect(wxml).toMatch(/微信支付/)
    expect(wxss).toMatch(/\.pay-method/)
  })

  it('福利金 is a discount toggle (switch), not a payment option', () => {
    expect(wxml).toMatch(/使用福利金抵扣/)
    expect(wxml).toMatch(/<switch[^>]*checked="\{\{useReward\}\}"/)
  })

  it('drops the redundant standalone 福利金 balance card (folded into the toggle desc)', () => {
    expect(wxml).not.toMatch(/reward-amount-card/)
  })
})

describe('order-confirm — 对齐 mock-screens 设计细节', () => {
  it('抵扣行 label + 金额都用深粉（设计稿；且 #FF4896 -¥5 对比度失守，深粉 4.96:1 达标）', () => {
    // 标签也走 text-discount（与金额同深粉），不再用亮粉 text-primary
    expect(wxml).toMatch(/slot="title"\s+class="text-discount">福利金抵扣/)
    expect(wxss).toMatch(/\.text-discount\s*\{[^}]*color:\s*var\(--green-2\)/)
  })

  it('微信徽章用新版微信绿 #07C160（设计稿）', () => {
    expect(wxss).toMatch(/\.wx-badge\s*\{[^}]*background:\s*#07C160/i)
    expect(wxss).not.toMatch(/#1AAD19/i)
  })
})
