# Implementation Plan: 配色改版 — 暖橙方案（取自 ui-add-food.png）

## Overview
当前 teal 主色（取自 ui-color.png）观感不佳。改为 `ui-add-food.png` 的**暖橙**配色：单一品牌橙 `#F88818`（选中分类/选规格徽标/底部选中 Tab 都用它）、顶部金黄→桃橙渐变、深青墨文字 `#083038`、白/浅灰中性底。由于全站 UI 已是 token 驱动（颜色集中在 `frontend/app.wxss` 的 `page{}`，组件只引用变量），本次改版主要是**改 token 值 + 少量硬编码 hex（导航栏 JSON、TabBar 兜底）+ 同步测试**，风险低。

颜色取样自 `ui-add-food.png`：强调橙 `#F88818`(248,136,24)、金黄 `#F0D060`、桃橙 `#F8A080`、深墨 `#083038`(8,48,56)。

## Architecture Decisions
- **单一品牌橙**：图中只有一种强调橙，故 `--brand-primary` 与 `--brand-accent` 同为 `#F88818`（保留两个 token 以兼容现有组件引用），按下态 `#E0760A`。
- **价格沿用深墨**（`--color-price: #083038`）以**忠于图片**（图中价格为深色，非橙色）。可一键改回橙色（见 Open Q1）。
- **新增 `--brand-gradient`**（金黄→桃橙）用于顶部 banner（启动页 hero、菜单门店栏），贴合图片顶部观感；其余渐变（原 primary→primary-dark）随 token 自动变暖橙。
- **导航栏色**只能写在各页 JSON（不支持 CSS 变量），故 5 个 JSON 的 `#189CA8` 需逐个改为 `#F88818`；`navigationBarTextStyle: white` 在橙底上可读，保持不变。
- **不改布局/结构/交互**，纯换色；DESIGN.md 仍是唯一真源。

## 目标调色板（写入 DESIGN.md / app.wxss）
```
品牌
--brand-primary        #F88818
--brand-primary-dark   #E0760A
--brand-primary-light  rgba(248,136,24,0.12)
--brand-accent         #F88818
--brand-accent-dark    #E0760A
--brand-accent-light   rgba(248,136,24,0.12)
--brand-gradient       linear-gradient(135deg, #F6C84A 0%, #F89A70 100%)
文本
--color-text-primary   #083038
--color-text-secondary rgba(8,48,56,0.60)
--color-text-hint      rgba(8,48,56,0.35)
--color-text-disabled  rgba(8,48,56,0.20)
--color-text-on-brand  #FFFFFF
中性/功能
--color-bg-page        #F5F5F5
--color-bg-surface     #FFFFFF
--color-bg-rail        #F5F5F5
--color-bg-hover       #FAF2E8
--color-border         rgba(8,48,56,0.08)
--color-border-light   rgba(8,48,56,0.04)
--color-price          #083038   (忠于图片；橙色为备选)
--color-success        #2BA471
--color-danger         #E5484D
weui 覆盖
--weui-primary         #F88818
--weui-primary-light   rgba(248,136,24,0.12)
--weui-color-success   #2BA471
```

## Dependency Graph
```
ui-add-food.png 取样
   └── DESIGN.md（配色真源）
          └── app.wxss token 值（组件已引用变量，无需改组件）
                 ├── 硬编码 hex：app.json + 4 页 JSON 导航栏色；tab-bar 兜底 hex
                 └── __tests__/design-tokens.test.js 断言（新 hex + 无残留 #189CA8）
          └── --brand-gradient → 启动页 hero / 菜单门店栏（贴合图片顶部）
```

---

## Task List

### Phase 1: 暖橙配色改版

#### Task 1: 重写 DESIGN.md 配色为暖橙
**Description:** 按上表把 DESIGN.md 第 1 节（配色）改为暖橙方案，更新「迁移说明」「价格」「分类栏」「TabBar」等描述与图片一致（来源由 ui-color.png 改为 ui-add-food.png），新增 `--brand-gradient`。仅改文档。
**Acceptance criteria:**
- [ ] DESIGN.md 配色表为上表暖橙值，无 `#189CA8` 残留。
- [ ] 标注价格用深墨（忠于图片）+ 橙色备选；说明单一品牌橙。
**Verification:**
- [ ] 人工通读 DESIGN.md 与本计划调色板一致。
**Dependencies:** None
**Files likely touched:** `DESIGN.md`
**Estimated scope:** S

#### Task 2: 应用 token 值 + 导航栏/TabBar hex + 测试
**Description:** 把 `app.wxss` 的 token 值改为暖橙（含新增 `--brand-gradient`）；将 `app.json` 与 4 个页面 JSON 的 `navigationBarBackgroundColor` 改为 `#F88818`；TabBar 兜底 hex `var(--brand-primary, #189CA8)` → `#F88818`；更新 `design-tokens.test.js` 断言为新 hex，并新增「无残留 `#189CA8`」扫描。
**Acceptance criteria:**
- [ ] `app.wxss` token 值与目标调色板一致；新增 `--brand-gradient`。
- [ ] 5 处导航栏色 = `#F88818`；TabBar 兜底 = `#F88818`。
- [ ] 全仓业务代码（排除 weui 库、废弃 tab-bar）无 `#189CA8`。
**Verification:**
- [ ] `cd frontend && node node_modules/jest/bin/jest.js` 全绿。
- [ ] `grep 189CA8 frontend`（排除 weui 库）无业务命中。
**Dependencies:** Task 1
**Files likely touched:** `frontend/app.wxss`, `frontend/app.json`, `frontend/pages/{home,menu,profile,invite,order-confirm}/index.json`, `frontend/miniprogram_npm/custom-tab-bar-comp/index.wxss`, `frontend/__tests__/design-tokens.test.js`
**Estimated scope:** M

#### Task 3: 暖色 banner 渐变贴合图片顶部
**Description:** 启动页 hero 与菜单门店栏使用 `--brand-gradient`（金黄→桃橙），更贴近图片顶部；其余组件随 token 自动变暖橙。可选：左侧选中分类底改用 `--brand-primary-light` 暖底（更贴近图片填充观感）。
**Acceptance criteria:**
- [ ] 启动页 hero 用 `--brand-gradient`。
- [ ] 菜单门店栏顶部呈暖色（gradient 或暖底），与图片观感一致。
**Verification:**
- [ ] jest 全绿；人工 DevTools 预览首页/菜单顶部为暖色。
**Dependencies:** Task 2
**Files likely touched:** `frontend/pages/home/index.wxss`, `frontend/pages/menu/index.wxss`
**Estimated scope:** S

#### Checkpoint: 配色改版完成
- [ ] jest 全绿；无残留 `#189CA8`（业务代码）。
- [ ] ⚠️ 人工在微信开发者工具确认首页/菜单/订单确认/我的等观感为暖橙且可读（白字在橙底、深墨文字）。
- [ ] 与人工确认「好看」后定稿；如需微调橙色明度/价格颜色按 Open Q 处理。

---

## Risks and Mitigations
| Risk | Impact | Mitigation |
|------|--------|------------|
| 导航栏白字在橙底对比度不足 | Low | `#F88818` 上白字对比度足够；Checkpoint 人工核验 |
| 漏改某处 `#189CA8` 造成冷暖混搭 | Med | Task 2 新增「无残留 #189CA8」测试扫描 + grep 验证 |
| 价格深墨 vs 橙色取向分歧 | Low | 默认忠于图片（深墨），Open Q1 一键切换 |
| 渐变仅在 banner，过度使用显廉价 | Low | `--brand-gradient` 仅用于顶部 banner，弱使用 |

## Open Questions
1. **价格颜色**：忠于图片用深墨 `#083038`，还是改回橙色 `#F88818` 更醒目？（计划默认深墨）
2. **品牌橙明度**：`#F88818` 直接取样；是否需要更柔和/更深一版？
3. **成功色**：由 teal 改为绿色 `#2BA471`（语义更准）是否可接受？

## Not Doing（及原因）
- **不改布局/交互/组件结构** —— 本次纯换色。
- **不动废弃 `custom-tab-bar/tab-bar/`** —— 未被引用（沿用既有决定）。
- **不改管理后台(Vue)视觉** —— 需求针对小程序。
