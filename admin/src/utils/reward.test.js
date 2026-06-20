import { describe, it, expect } from 'vitest'
import { toPercent, toDecimal, parseCategories } from './reward'

describe('toPercent (decimal -> percent for display)', () => {
  it('converts a rate to a percent', () => {
    expect(toPercent(0.03)).toBe(3)
    expect(toPercent(0.1)).toBe(10)
    expect(toPercent(0.5)).toBe(50)
  })
  it('keeps fractional percents', () => {
    expect(toPercent(0.035)).toBe(3.5)
  })
  it('treats 0 as 0 (not dropped)', () => {
    expect(toPercent(0)).toBe(0)
  })
  it('treats null/undefined as 0', () => {
    expect(toPercent(null)).toBe(0)
    expect(toPercent(undefined)).toBe(0)
  })
  it('avoids float dust', () => {
    // 0.03 * 100 === 3.0000000000000004 without rounding
    expect(toPercent(0.03)).toBe(3)
  })
})

describe('toDecimal (percent -> decimal for storage)', () => {
  it('converts a percent to a rate', () => {
    expect(toDecimal(3)).toBe(0.03)
    expect(toDecimal(10)).toBe(0.1)
    expect(toDecimal(50)).toBe(0.5)
  })
  it('keeps fractional rates', () => {
    expect(toDecimal(3.5)).toBe(0.035)
  })
  it('treats 0 as 0', () => {
    expect(toDecimal(0)).toBe(0)
  })
  it('treats null/undefined as 0', () => {
    expect(toDecimal(null)).toBe(0)
    expect(toDecimal(undefined)).toBe(0)
  })
})

describe('round-trip', () => {
  it('toDecimal(toPercent(x)) === x for typical rates', () => {
    for (const x of [0, 0.03, 0.04, 0.1, 0.12, 0.5]) {
      expect(toDecimal(toPercent(x))).toBe(x)
    }
  })
})

describe('parseCategories (jsonb array string -> array)', () => {
  it('parses a valid array', () => {
    expect(parseCategories('["热菜","饮品"]')).toEqual(['热菜', '饮品'])
  })
  it('returns [] for empty/null', () => {
    expect(parseCategories('')).toEqual([])
    expect(parseCategories(null)).toEqual([])
    expect(parseCategories('[]')).toEqual([])
  })
  it('returns [] for invalid JSON', () => {
    expect(parseCategories('not json')).toEqual([])
  })
  it('returns [] when JSON is not an array', () => {
    expect(parseCategories('{"a":1}')).toEqual([])
    expect(parseCategories('42')).toEqual([])
  })
})
