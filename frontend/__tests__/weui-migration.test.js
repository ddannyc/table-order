/**
 * Enforces the weui de-customization: migrated pages must not reference the
 * LEGACY parallel color system (--color-*, the removed --weui-primary, or
 * stray --brand-* tokens). They should use weui's own color vars (--weui-FG-*,
 * --weui-BG-*, --weui-BRAND) instead. Non-color scale tokens (--space/--font/
 * --radius) are allowed to remain.
 *
 * Sanctioned exceptions (松墨 Pine-Ink texture uplift, see
 * docs/ideas/texture-uplift-pine-ink.md): --brand-accent (gold) and --price-ink
 * are intentional palette tokens defined in app.wxss and may be referenced.
 *
 * Add each page to `migrated` as its task lands (RED until migrated).
 */
const fs = require('fs')
const path = require('path')

const read = (rel) => fs.readFileSync(path.join(__dirname, '..', rel), 'utf8')
// Block legacy --color-*/--weui-primary and any --brand-* EXCEPT --brand-accent.
const CUSTOM_COLOR = /var\(--(?:brand-(?!accent)|color-|weui-primary)/

const migrated = ['login', 'home', 'menu', 'order-confirm', 'profile', 'invite', 'share-code']

describe('migrated pages use no custom color tokens', () => {
  it.each(migrated)('%s/index.wxss has no --brand/--color/--weui-primary', (p) => {
    expect(read(`pages/${p}/index.wxss`)).not.toMatch(CUSTOM_COLOR)
  })
})
