/**
 * Navigation bars follow the v6 cream/pine scheme so the top area shares the
 * unified cream base (#F3EEE4). Home keeps a pine bar (#234B3A) so its green
 * brand header reaches the top; the other pages use the cream top with dark text.
 */
const fs = require('fs')
const path = require('path')
const j = (rel) => JSON.parse(fs.readFileSync(path.join(__dirname, '..', rel), 'utf8'))

describe('navigation bars — v6 cream/pine scheme', () => {
  it('app.json window defaults to the cream base with dark text', () => {
    const w = j('app.json').window
    expect(w.navigationBarBackgroundColor).toBe('#F3EEE4')
    expect(w.navigationBarTextStyle).toBe('black')
  })

  it('home uses a brand-pink bar so the brand header reaches the top', () => {
    const h = j('pages/home/index.json')
    expect(h.navigationBarBackgroundColor).toBe('#FF4896')
    expect(h.navigationBarTextStyle).toBe('white')
  })

  it.each(['menu', 'invite', 'profile', 'order-confirm'])('%s uses the cream top', (p) => {
    const c = j(`pages/${p}/index.json`)
    expect(c.navigationBarBackgroundColor).toBe('#F3EEE4')
    expect(c.navigationBarTextStyle).toBe('black')
  })

  it('no leftover bright weui green (#07c160) anywhere', () => {
    expect(j('app.json').window.navigationBarBackgroundColor).not.toBe('#07c160')
    ;['home', 'menu', 'invite', 'profile', 'order-confirm'].forEach((p) => {
      expect(j(`pages/${p}/index.json`).navigationBarBackgroundColor).not.toBe('#07c160')
    })
  })
})
