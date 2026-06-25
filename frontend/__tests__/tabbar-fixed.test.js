/**
 * mp-tabbar must be pinned to the screen bottom (Task 1).
 * It is not position:fixed by default, so each tab page wraps it in a
 * global .tabbar-fixed container.
 */
const fs = require('fs')
const path = require('path')
const read = (rel) => fs.readFileSync(path.join(__dirname, '..', rel), 'utf8')

describe('tabbar pinned to bottom', () => {
  const wxss = read('app.wxss')
  it('app.wxss defines .tabbar-fixed as fixed/bottom', () => {
    expect(wxss).toMatch(/\.tabbar-fixed\s*\{[^}]*position:\s*fixed/)
    expect(wxss).toMatch(/\.tabbar-fixed\s*\{[^}]*bottom:\s*0/)
  })

  it.each(['home', 'menu', 'invite', 'profile'])('%s wraps mp-tabbar in .tabbar-fixed', (p) => {
    const wxml = read(`pages/${p}/index.wxml`)
    expect(wxml).toMatch(/tabbar-fixed[^>]*>\s*<mp-tabbar/)
  })
})
