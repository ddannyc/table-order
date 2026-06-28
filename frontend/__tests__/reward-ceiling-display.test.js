/**
 * 福利金抵扣上限的展示必须读取商家设置值，不在小程序里写死百分比。
 * - 下单页（order-confirm）有店铺上下文：绑定 {{rewardCeiling}}（由 shop.reward_ceiling 计算）。
 * - 我的/邀请页无店铺上下文：改为不含具体百分比的通用文案（避免写死且跨店误导）。
 */
const fs = require('fs')
const path = require('path')
const read = (p) => fs.readFileSync(path.join(__dirname, '..', p), 'utf8')

describe('福利金抵扣上限展示不写死、读设置值', () => {
  it('下单页抵扣上限来自 shop.reward_ceiling 的绑定值', () => {
    const wxml = read('pages/order-confirm/index.wxml')
    expect(wxml).toMatch(/最高抵扣\s*\{\{rewardCeiling\}\}%/)
    const js = read('pages/order-confirm/index.js')
    expect(js).toMatch(/shop\.reward_ceiling/)
  })

  it('邀请页抵扣文案不写死百分比', () => {
    const wxml = read('pages/invite/index.wxml')
    expect(wxml).not.toMatch(/抵扣[^<]*\d+%/)
  })

  it('我的页抵扣上限不写死百分比', () => {
    const wxml = read('pages/profile/index.wxml')
    expect(wxml).not.toMatch(/抵扣上限[^<]*\d+%/)
  })
})
