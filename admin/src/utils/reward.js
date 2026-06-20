// Reward rates are stored as decimals (0.03) but shown/edited as percents (3%).

export const toPercent = (decimal) => Number(((decimal || 0) * 100).toFixed(2))

export const toDecimal = (percent) => Number(((percent || 0) / 100).toFixed(4))

// reward_exclude_categories is a jsonb array string, e.g. '["热菜"]'.
export function parseCategories(raw) {
  try {
    const arr = JSON.parse(raw || '[]')
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}
