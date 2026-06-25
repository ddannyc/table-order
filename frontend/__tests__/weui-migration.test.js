/**
 * Enforces the weui de-customization: migrated pages must not reference custom
 * COLOR tokens (--brand-*, --color-*) or the removed --weui-primary override.
 * They should use weui components/classes and weui's own color vars (--weui-FG-*,
 * --weui-BG-*, --weui-BRAND) instead. Non-color scale tokens (--space/--font/
 * --radius) are allowed to remain.
 *
 * Add each page to `migrated` as its task lands (RED until migrated).
 */
const fs = require('fs')
const path = require('path')

const read = (rel) => fs.readFileSync(path.join(__dirname, '..', rel), 'utf8')
const CUSTOM_COLOR = /var\(--(brand-|color-|weui-primary)/

const migrated = ['login', 'home']

describe('migrated pages use no custom color tokens', () => {
  it.each(migrated)('%s/index.wxss has no --brand/--color/--weui-primary', (p) => {
    expect(read(`pages/${p}/index.wxss`)).not.toMatch(CUSTOM_COLOR)
  })
})
