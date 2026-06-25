/**
 * Navigation bars use the weui default green (Task 2) — no leftover orange.
 */
const fs = require('fs')
const path = require('path')
const j = (rel) => JSON.parse(fs.readFileSync(path.join(__dirname, '..', rel), 'utf8'))

describe('navigation bar uses weui green #07c160', () => {
  it('app.json window', () => {
    expect(j('app.json').window.navigationBarBackgroundColor).toBe('#07c160')
  })
  it.each(['home', 'menu', 'invite', 'profile', 'order-confirm'])('%s page', (p) => {
    expect(j(`pages/${p}/index.json`).navigationBarBackgroundColor).toBe('#07c160')
  })
})
