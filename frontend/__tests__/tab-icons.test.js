/**
 * Tab icon color guard (D5).
 * The active tab icons must be the BRAND pine green (#234B3A), matching the
 * v6 design token — not the old bright weui green (#07C160). Decodes each PNG
 * and scans pixels, so a regression to the wrong green fails the suite.
 */
const fs = require('fs')
const path = require('path')
const { inflateSync } = require('zlib')

function paeth(a, b, c) {
  const p = a + b - c
  const pa = Math.abs(p - a)
  const pb = Math.abs(p - b)
  const pc = Math.abs(p - c)
  return pa <= pb && pa <= pc ? a : pb <= pc ? b : c
}

function decodePNG(file) {
  const buf = fs.readFileSync(file)
  const w = buf.readUInt32BE(16)
  const h = buf.readUInt32BE(20)
  let p = 8
  const idat = []
  while (p < buf.length) {
    const len = buf.readUInt32BE(p)
    const type = buf.toString('ascii', p + 4, p + 8)
    if (type === 'IDAT') idat.push(buf.slice(p + 8, p + 8 + len))
    p += 12 + len
  }
  const raw = inflateSync(Buffer.concat(idat))
  const bpp = 4
  const stride = w * bpp
  const out = Buffer.alloc(h * stride)
  for (let y = 0; y < h; y++) {
    const ft = raw[y * (stride + 1)]
    const base = y * (stride + 1) + 1
    for (let x = 0; x < stride; x++) {
      const a = x >= bpp ? out[y * stride + x - bpp] : 0
      const b = y > 0 ? out[(y - 1) * stride + x] : 0
      const c = x >= bpp && y > 0 ? out[(y - 1) * stride + x - bpp] : 0
      let v = raw[base + x]
      if (ft === 1) v = (v + a) & 255
      else if (ft === 2) v = (v + b) & 255
      else if (ft === 3) v = (v + ((a + b) >> 1)) & 255
      else if (ft === 4) v = (v + paeth(a, b, c)) & 255
      out[y * stride + x] = v
    }
  }
  return { w, h, data: out }
}

function hasColor({ w, h, data }, [R, G, B], tol = 24) {
  for (let i = 0; i < w * h; i++) {
    if (data[i * 4 + 3] < 180) continue
    const dr = data[i * 4] - R
    const dg = data[i * 4 + 1] - G
    const db = data[i * 4 + 2] - B
    if (dr * dr + dg * dg + db * db <= tol * tol * 3) return true
  }
  return false
}

describe('tab icons — BRAND pine-green recolor (D5)', () => {
  it.each(['menu', 'invite', 'profile'])(
    '%s-active icon is BRAND green (#234B3A), not bright weui green',
    (name) => {
      const img = decodePNG(path.join(__dirname, `../static/${name}-active.png`))
      expect(hasColor(img, [35, 75, 58], 8)).toBe(true) // BRAND #234B3A present (tight: pixels are uniform)
      expect(hasColor(img, [7, 193, 96])).toBe(false) // no bright weui green left
    }
  )
})
