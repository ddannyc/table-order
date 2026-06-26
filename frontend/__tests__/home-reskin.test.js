/**
 * Structural tests for the 松墨 Pine-Ink home reskin (T2/T3).
 * Reads the page's wxml/wxss as text and asserts the brand band, token-based
 * surfaces, and (T3) the hero illustration are present. Behavior is covered
 * separately by home-launcher.test.js — this file only guards the skin.
 */
const fs = require('fs')
const path = require('path')

const wxml = fs.readFileSync(path.join(__dirname, '../pages/home/index.wxml'), 'utf8')
const wxss = fs.readFileSync(path.join(__dirname, '../pages/home/index.wxss'), 'utf8')

describe('home reskin — brand band + entry cards (T2)', () => {
  it('wraps the header in a deep-green brand band', () => {
    expect(wxml).toMatch(/home-brandband/)
    expect(wxss).toMatch(/\.home-brandband\s*\{[^}]*background:\s*var\(--weui-BRAND\)/)
  })

  it('entry cards use the warm card surface token, not hardcoded white', () => {
    expect(wxss).toMatch(/\.entry-card\s*\{[^}]*background:\s*var\(--weui-BG-1\)/)
    expect(wxss).not.toMatch(/\.entry-card\s*\{[^}]*background:\s*#fff/)
  })

  it('entry titles use the brand green', () => {
    expect(wxss).toMatch(/\.entry-name\s*\{[^}]*color:\s*var\(--weui-BRAND\)/)
  })
})

describe('home reskin — hero illustration (T3)', () => {
  it('renders a single-line SVG hero banner (zero binary assets)', () => {
    expect(wxml).toMatch(/home-hero/)
    expect(wxss).toMatch(/\.home-hero\s*\{[^}]*data:image\/svg\+xml/)
  })

  it('respects prefers-reduced-motion if the hero animates', () => {
    // If a transition/animation is added to the hero, it must be disabled
    // under reduced motion. (No animation is also acceptable.)
    const animatesHero = /\.home-hero\s*\{[^}]*(animation|transition)/.test(wxss)
    if (animatesHero) {
      expect(wxss).toMatch(/prefers-reduced-motion/)
    }
  })
})
