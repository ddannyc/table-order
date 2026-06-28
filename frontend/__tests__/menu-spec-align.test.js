/**
 * 校验点餐页对齐 mock-spec.html 的版式细节（无需真实图/促销数据的项）。
 * 颜色令牌的对比度由 theme-tokens.test.js 把关；本文件只守护结构/字重/版式。
 */
const fs = require('fs')
const path = require('path')

const wxss = fs.readFileSync(path.join(__dirname, '../pages/menu/index.wxss'), 'utf8')
const wxml = fs.readFileSync(path.join(__dirname, '../pages/menu/index.wxml'), 'utf8')
const rule = (sel) => {
  const m = wxss.match(new RegExp('\\' + sel + '\\s*\\{([^}]*)\\}'))
  return m ? m[1] : null
}

describe('menu — # 板块标题加蓝色横线（设计签名）', () => {
  it('结构：标题文字 + 蓝线条', () => {
    expect(wxml).toMatch(/menu-sec-t/)
    expect(wxml).toMatch(/menu-sec-line/)
  })
  it('# 前缀在标题文字上、蓝线填充整行', () => {
    expect(wxss).toMatch(/\.menu-sec-t::before\s*\{[^}]*content:\s*"#"/)
    const ln = rule('.menu-sec-line')
    expect(ln).toMatch(/flex:\s*1/)
    expect(ln).toMatch(/background:\s*var\(--jf-blue\)/)
  })
})

describe('menu — 黄色药丸内必须深色字（防白字 on 黄不可读回归）', () => {
  it('左轨数量徽章用深蓝字，绝不用白字（白 on 黄仅 1.44:1）', () => {
    const c = rule('.menu-rail-count')
    expect(c).toMatch(/color:\s*var\(--jf-title-blue\)/)
    expect(c).not.toMatch(/color:\s*#fff/i)
  })
})

describe('menu — 门店栏对齐', () => {
  it('门店栏透明（坐在页面粉底上，非白底）', () => {
    expect(rule('.menu-shopbar')).not.toMatch(/background:\s*var\(--weui-BG-1\)/)
  })
  it('定位针是粉色实心圆（白针在内）', () => {
    const p = rule('.menu-shop-pin')
    expect(p).toMatch(/background-color:\s*var\(--weui-BRAND\)/)
    expect(p).toMatch(/border-radius:\s*50%/)
  })
  it('店名 900', () => {
    expect(rule('.menu-shopname')).toMatch(/font-weight:\s*900/)
  })
  it('点餐方式药丸：浅粉底 + 深粉字（非实心粉+白）', () => {
    const m = rule('.menu-mode')
    expect(m).toMatch(/background:\s*var\(--jf-tag-pink\)/)
    expect(m).toMatch(/color:\s*var\(--jf-tag-red\)/)
  })
})

describe('menu — 卡片字重 + 价格标 3D', () => {
  it('菜名 900', () => {
    expect(rule('.menu-card-title')).toMatch(/font-weight:\s*900/)
  })
  it('价格标 800 + 黄色立体底影', () => {
    const p = rule('.menu-price')
    expect(p).toMatch(/font-weight:\s*800/)
    expect(p).toMatch(/box-shadow:[^;]*#E0B900/i)
  })
  it('加减器是粉边圆按钮（非 weui 灰按钮）', () => {
    expect(wxml).toMatch(/menu-step-btn/)
    expect(wxml).not.toMatch(/menu-stepper">[\s\S]*weui-btn_default/)
    const b = rule('.menu-step-btn')
    expect(b).toMatch(/border:[^;]*var\(--weui-BRAND\)/)
    expect(b).toMatch(/border-radius:\s*50%/)
  })
})

describe('menu — 品牌标题 + 购物车浮动药丸', () => {
  it('品牌标题加深粉立体投影', () => {
    expect(rule('.menu-poster-brand')).toMatch(/text-shadow:[^;]*var\(--green-2\)/)
  })
  it('购物车条是浮动药丸：内缩 + 大圆角 + 实心粉（非满宽渐变）', () => {
    const c = rule('.menu-cartbar')
    expect(c).toMatch(/left:\s*var\(--space/)
    expect(c).toMatch(/border-radius/)
    expect(c).not.toMatch(/linear-gradient/)
  })
  it('车标圆是黄底', () => {
    expect(rule('.menu-cart-icon')).toMatch(/background:\s*var\(--accent\)/)
  })
  it('去结算 900', () => {
    expect(rule('.menu-cart-go')).toMatch(/font-weight:\s*900/)
  })
})
