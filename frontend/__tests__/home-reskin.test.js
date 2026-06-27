/**
 * Structural tests for the v6 home (D2).
 * Home is: brand header (NO balance/reward) + 堂食/外卖 segmented (line icons)
 * + 扫码点餐 card + chef banner + 福利放送 static promo cards.
 * Behavior (scan / delivery) is covered by home-launcher.test.js.
 */
const fs = require('fs')
const path = require('path')

const read = (p) => fs.readFileSync(path.join(__dirname, p), 'utf8')
const wxml = read('../pages/home/index.wxml')
const wxss = read('../pages/home/index.wxss')
const art = read('../pages/home/home-art.wxss')

describe('home v6 — brand header (no wallet)', () => {
  it('shows the brand header, not a wallet balance/reward', () => {
    expect(wxml).toMatch(/home-hd/)
    expect(wxml).toMatch(/鸡福旺/)
    expect(wxml).not.toMatch(/balanceText/)
    expect(wxml).not.toMatch(/rewardText/)
  })
})

describe('home v6 — segmented 堂食/外卖 with line icons', () => {
  it('renders both segments with line-icon classes and the right actions', () => {
    expect(wxml).toMatch(/seg-ic_dine/)
    expect(wxml).toMatch(/seg-ic_deliver/)
    expect(wxml).toMatch(/bindtap="selectDineIn"/)
    expect(wxml).toMatch(/bindtap="chooseDelivery"/)
  })

  it('the segment icons recolor between states (white inactive, pink active)', () => {
    expect(art).toMatch(/\.seg-ic_dine\s*\{[^}]*data:image\/svg\+xml/)
    expect(art).toMatch(/s_on .seg-ic_dine\s*\{[^}]*%23FF4896/i)
  })
})

describe('home v6 — scan card + chef banner + promos', () => {
  it('scan card triggers scanDineIn', () => {
    expect(wxml).toMatch(/scan/)
    expect(wxml).toMatch(/bindtap="scanDineIn"/)
  })

  it('hero banner is a colored data-uri illustration (pink outline + orange fills, no Pine-Ink residue)', () => {
    expect(wxml).toMatch(/home-hero/)
    expect(art).toMatch(/\.home-hero\s*\{[^}]*data:image\/svg\+xml/)
    expect(art).toMatch(/%23FF4896/i)
    expect(art).toMatch(/%23F0801A/i)
    expect(art).not.toMatch(/%23234B3A/i)
    expect(art).not.toMatch(/%23C8643C/i)
  })

  it('福利放送 has two static illustrated promo cards', () => {
    expect(wxml).toMatch(/promo-bowl/)
    expect(wxml).toMatch(/promo-greens/)
    expect(art).toMatch(/\.promo-bowl\s*\{[^}]*data:image\/svg\+xml/)
  })
})

describe('home — 对齐 mock-screens 设计细节', () => {
  const rule = (sel) => wxss.match(new RegExp('\\' + sel + '\\s*\\{([^}]*)\\}'))[1]

  it('logo 用黄底（品牌字标，对齐设计稿）', () => {
    expect(rule('.logo')).toMatch(/background:\s*var\(--accent\)/)
  })

  it('“更多›”用蓝色点缀 + 加粗（设计稿 sect-more 是蓝色）', () => {
    const m = rule('.sec-more')
    expect(m).toMatch(/color:\s*var\(--jf-blue\)/)
    expect(m).toMatch(/font-weight:\s*700/)
  })

  it('promo 菜名加粗 800（设计稿 promo-n 800）', () => {
    expect(rule('.promo-n')).toMatch(/font-weight:\s*800/)
  })

  it('扫码图标框用浅粉底 + 深粉二维码（设计稿 tag-pink 底 + tag-red 码）', () => {
    expect(rule('.scan-qr')).toMatch(/background-color:\s*var\(--jf-tag-pink\)/)
    expect(art).toMatch(/\.scan-qr-ic\s*\{[^}]*%23C2185B/i)
    expect(art).not.toMatch(/\.scan-qr-ic\s*\{[^}]*stroke%3D'%23fff'/i)
  })
})
