/**
 * 菜单加载态用骨架屏（Variant B）替代居中转圈，缓解外卖冷启动的「白屏加载中」观感。
 * - 静态品牌头（#鸡福旺 / 现炸出炉）在 loading 态即刻渲染（不占位）。
 * - 仅动态区（店名 / 左类目轨 / 卡片）用 shimmer 占位，且复用真实布局类避免布局跳动。
 * - shimmer 动画尊重 prefers-reduced-motion。
 */
const fs = require('fs')
const path = require('path')
const read = (p) => fs.readFileSync(path.join(__dirname, '..', p), 'utf8')

describe('菜单加载骨架屏（Variant B）', () => {
  const wxml = read('pages/menu/index.wxml')
  // 截取 loading 分支区块（wx:elif loading → 下一个 wx:elif error 之间）
  const loadingBlock = wxml.split('wx:elif="{{loading}}"')[1].split('wx:elif="{{error}}"')[0]

  it('loading 态不再用 weui-loadmore 居中转圈', () => {
    expect(wxml).not.toMatch(/weui-loadmore/)
  })

  it('loading 态渲染骨架屏容器', () => {
    expect(loadingBlock).toMatch(/menu-skeleton/)
  })

  it('静态品牌头在骨架态即刻出现（Variant B：不占位）', () => {
    expect(loadingBlock).toMatch(/#鸡福旺/)
    expect(loadingBlock).toMatch(/现炸出炉/)
  })

  it('卡片骨架复用真实布局尺寸（缩略图块 + shimmer 卡片）', () => {
    expect(loadingBlock).toMatch(/sk-card/)
    expect(loadingBlock).toMatch(/menu-card-thumb/)
  })

  it('shimmer 动画存在且尊重 reduced-motion', () => {
    const wxss = read('pages/menu/index.wxss')
    expect(wxss).toMatch(/@keyframes\s+sk-shimmer/)
    expect(wxss).toMatch(/prefers-reduced-motion/)
  })
})
