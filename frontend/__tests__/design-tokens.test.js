/**
 * Tests for DESIGN.md brand tokens — warm-orange palette (#F88818) from
 * ui-add-food.png.
 *
 * Verifies app.wxss tokens and per-page navigation bars use the warm orange
 * and that the prior teal (#189CA8) and WeChat green (#07c160) are gone.
 */
const fs = require('fs')
const path = require('path')

const wxss = fs.readFileSync(path.join(__dirname, '..', 'app.wxss'), 'utf8')
const appJson = JSON.parse(fs.readFileSync(path.join(__dirname, '..', 'app.json'), 'utf8'))

// pull a CSS var value out of app.wxss, e.g. --brand-primary: #F88818;
const tokenValue = (name) => {
  const m = wxss.match(new RegExp(`${name}\\s*:\\s*([^;]+);`))
  return m ? m[1].trim() : null
}

describe('DESIGN.md brand tokens in app.wxss', () => {
  it('defines the warm orange brand primary token', () => {
    expect(tokenValue('--brand-primary')).toBe('#F88818')
  })

  it('defines the brand accent as the same warm orange', () => {
    expect(tokenValue('--brand-accent')).toBe('#F88818')
  })

  it('defines brand primary dark + light variants', () => {
    expect(tokenValue('--brand-primary-dark')).toBe('#E0760A')
    expect(tokenValue('--brand-primary-light')).toBe('rgba(248,136,24,0.12)')
  })

  it('defines the warm header gradient token', () => {
    expect(tokenValue('--brand-gradient')).toMatch(/#F6C84A.*#F89A70/)
  })

  it('keeps price on the deep-ink color (faithful to image)', () => {
    expect(tokenValue('--color-price')).toBe('#083038')
  })

  it('migrates --weui-primary to the warm orange', () => {
    expect(tokenValue('--weui-primary')).toBe('#F88818')
  })

  it('leaves no hardcoded WeChat green or teal in app.wxss', () => {
    expect(wxss).not.toMatch(/#07c160/i)
    expect(wxss).not.toMatch(/#189CA8/i)
  })
})

describe('app.json navigation bar', () => {
  it('uses the warm orange for the navigation bar', () => {
    expect(appJson.window.navigationBarBackgroundColor).toBe('#F88818')
  })
})

describe('no residual teal or green in business styles', () => {
  const businessFiles = [
    'miniprogram_npm/custom-tab-bar-comp/index.wxss',
    'pages/invite/index.wxss',
    'pages/profile/index.wxss',
    'pages/share-code/index.wxss',
    'pages/home/index.json',
    'pages/menu/index.json',
    'pages/invite/index.json',
    'pages/profile/index.json',
    'pages/order-confirm/index.json',
  ]

  it.each(businessFiles)('%s contains no #07c160 / #189CA8', (rel) => {
    const content = fs.readFileSync(path.join(__dirname, '..', rel), 'utf8')
    expect(content).not.toMatch(/#07c160/i)
    expect(content).not.toMatch(/#059a4c/i) // green gradient tail
    expect(content).not.toMatch(/#189CA8/i) // prior teal
  })

  it('custom tab-bar active state references the brand token', () => {
    const css = fs.readFileSync(
      path.join(__dirname, '..', 'miniprogram_npm/custom-tab-bar-comp/index.wxss'),
      'utf8'
    )
    expect(css).toMatch(/var\(--brand-primary/)
  })
})
