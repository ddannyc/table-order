/**
 * Tests for DESIGN.md brand tokens — global reskin (teal #189CA8 + orange #FC8400).
 *
 * Verifies app.wxss design tokens and app.json navigation bar are migrated off
 * WeChat green (#07c160) to the palette in DESIGN.md.
 */
const fs = require('fs')
const path = require('path')

const wxss = fs.readFileSync(path.join(__dirname, '..', 'app.wxss'), 'utf8')
const appJson = JSON.parse(fs.readFileSync(path.join(__dirname, '..', 'app.json'), 'utf8'))

// pull a CSS var value out of app.wxss, e.g. --brand-primary: #189CA8;
const tokenValue = (name) => {
  const m = wxss.match(new RegExp(`${name}\\s*:\\s*([^;]+);`))
  return m ? m[1].trim() : null
}

describe('DESIGN.md brand tokens in app.wxss', () => {
  it('defines the teal brand primary token', () => {
    expect(tokenValue('--brand-primary')).toBe('#189CA8')
  })

  it('defines the orange brand accent token', () => {
    expect(tokenValue('--brand-accent')).toBe('#FC8400')
  })

  it('defines brand primary dark + light variants', () => {
    expect(tokenValue('--brand-primary-dark')).toBe('#0C7A85')
    expect(tokenValue('--brand-primary-light')).toBe('rgba(24,156,168,0.12)')
  })

  it('defines the orange price token', () => {
    expect(tokenValue('--color-price')).toBe('#FC8400')
  })

  it('migrates --weui-primary off WeChat green to teal', () => {
    expect(tokenValue('--weui-primary')).toBe('#189CA8')
  })

  it('leaves no hardcoded WeChat green in app.wxss', () => {
    expect(wxss).not.toMatch(/#07c160/i)
  })
})

describe('app.json navigation bar', () => {
  it('uses the teal brand primary for the navigation bar', () => {
    expect(appJson.window.navigationBarBackgroundColor).toBe('#189CA8')
  })
})
