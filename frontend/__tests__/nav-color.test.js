/**
 * Navigation bars follow the 鸡福旺 (JFW) scheme: the top area shares the light
 * page base (#FBEFF3). Home keeps a brand-pink bar (#FF4896) so its pink brand
 * header reaches the top; the other pages use the light top with dark text.
 */
const fs = require('fs')
const path = require('path')
const j = (rel) => JSON.parse(fs.readFileSync(path.join(__dirname, '..', rel), 'utf8'))

describe('navigation bars — JFW pink/light scheme', () => {
  it('app.json window defaults to the light page base with dark text', () => {
    const w = j('app.json').window
    expect(w.navigationBarBackgroundColor).toBe('#FBEFF3')
    expect(w.navigationBarTextStyle).toBe('black')
  })

  it('home uses a brand-pink bar so the brand header reaches the top', () => {
    const h = j('pages/home/index.json')
    expect(h.navigationBarBackgroundColor).toBe('#FF4896')
    expect(h.navigationBarTextStyle).toBe('white')
  })

  it.each(['menu', 'invite', 'profile', 'order-confirm'])('%s uses the light top', (p) => {
    const c = j(`pages/${p}/index.json`)
    expect(c.navigationBarBackgroundColor).toBe('#FBEFF3')
    expect(c.navigationBarTextStyle).toBe('black')
  })

  it('no leftover bright weui green (#07c160) anywhere', () => {
    expect(j('app.json').window.navigationBarBackgroundColor).not.toBe('#07c160')
    ;['home', 'menu', 'invite', 'profile', 'order-confirm'].forEach((p) => {
      expect(j(`pages/${p}/index.json`).navigationBarBackgroundColor).not.toBe('#07c160')
    })
  })
})
