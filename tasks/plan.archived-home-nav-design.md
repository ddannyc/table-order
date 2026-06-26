# Implementation Plan: 首页/导航 设计修复（weui 体系内）

## Overview
基于截图评审，修复 de-customization 后遗留的结构/一致性问题，全部在 weui 体系内，不重新引入定制品牌色。重点是两个 bug（TabBar 浮空、橙绿不一致）+ 首页观感（入口放大、外卖未上线态、图标尺寸）。

## Architecture Decisions
- **不加定制配色/渐变/插画**：用户已确定 weui 默认绿，保持克制。
- **TabBar 用 fixed 包裹**：mp-tabbar 默认非 fixed，包一层全局 `.tabbar-fixed` 钉底；`.page` 已有 `padding-bottom` 兜底。
- **导航栏统一 weui 绿 `#07c160`**：去掉橙色。
- **图标加内边距重生成**：当前图标紧贴边框显大，加 ~18% 透明边距即视觉变小、与文字协调。

## Task List

### Task 1: TabBar 钉底（bug，影响 4 页）
**Acceptance:** [ ] home/menu/invite/profile 的 mp-tabbar 包在 `.tabbar-fixed`（position:fixed,bottom:0）内；app.wxss 定义该类。 [ ] 页面底部不再有大块空白、tabbar 在屏底。
**Verification:** [ ] 测试：4 页 wxml 含 `tabbar-fixed` 包裹 mp-tabbar；app.wxss 含 `.tabbar-fixed{position:fixed`；jest 全绿。
**Files:** `frontend/app.wxss`, `frontend/pages/{home,menu,invite,profile}/index.wxml`
**Scope:** M

### Task 2: 导航栏橙→绿（一致性 bug，5 页）
**Acceptance:** [ ] app.json + home/menu/invite/profile/order-confirm 的 `navigationBarBackgroundColor` = `#07c160`。
**Verification:** [ ] 测试断言 6 处 nav 色为 `#07c160`；jest 全绿。
**Files:** `frontend/app.json`, `frontend/pages/{home,menu,invite,profile,order-confirm}/index.json`
**Scope:** S

### Task 3: 首页入口放大 + 外卖未上线态
**Acceptance:** [ ] 两入口为等高大卡（堂食上/外卖下），图标更大、信息居中、重心下移。 [ ] 外卖卡整体置灰，"即将上线"为独立徽标；点按给 toast（不进入未完成流程）。 [ ] 堂食 scan、外卖 toast 逻辑不变。
**Verification:** [ ] home wxml 含独立 `home-badge` 徽标 + 外卖卡 `is-soon`/disabled 标识；home-launcher 测试全绿。
**Files:** `frontend/pages/home/index.wxml`, `frontend/pages/home/index.wxss`
**Scope:** M

### Task 4: tab 图标尺寸（加内边距重生成）
**Acceptance:** [ ] 重新生成 menu/invite/profile(+active) 图标，含透明内边距（不再满框），视觉与文字协调。
**Verification:** [ ] 测试：图标 PNG 的 getbbox 小于整幅（存在边距）；jest 全绿。
**Files:** `frontend/static/{menu,invite,profile}(-active).png`
**Scope:** S

### Checkpoint
- [ ] jest 全绿；⚠️ 人工 DevTools 验收：tabbar 钉底、全站绿、首页入口平衡、图标协调。

## Risks
| Risk | Mitigation |
|------|------------|
| fixed tabbar 遮挡内容 | `.page` padding-bottom 已留；菜单 cartbar 在 tabbar 之上已处理 |
| 原生 tabBar vs 组件 | 本期用 fixed 包裹组件（成本低）；如需更稳可后续转原生 tabBar |

## Not Doing
- 不加定制色/渐变/动画（保持 weui 克制）。
- 不改后端/业务逻辑。
