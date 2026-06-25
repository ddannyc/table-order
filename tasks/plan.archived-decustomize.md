# Implementation Plan: 去定制化 — 全面改用 weui 原生组件 + 默认配色

## Overview
移除项目的定制设计体系（`DESIGN.md` + `app.wxss` 里的 `--brand-*` / `--color-*` token），所有页面改用 **weui 原生组件 / 类**（`weui-miniprogram` 已是依赖），配色回到 **weui 默认（微信绿 `#07c160`，仅浅色）**，不再覆盖任何 `--weui-*` 变量。底部导航换成 **weui tabbar 组件**，菜单页**重构为纯 weui 形态**（顶部 `weui-navbar` 分类页签 + weui 媒体列表，不再左右两栏）。

参考：weui-wxss 文档（context7 `/tencent/weui-wxss`）。可用 weui 组件：`tabbar / navigation-bar / searchbar / cells / cell / dialog / half-screen-dialog / grids / actionsheet / toptips / msg / loading / icon / form / form-page`。

## Architecture Decisions
- **配色基准 = weui 默认绿 `#07c160`，仅浅色**：不再定义/覆盖品牌色变量；组件用 weui 自带类样式（`weui-btn_primary` 等）。
- **菜单重构为纯 weui 形态**：顶部 `weui-navbar` 分类页签 + 每分类一个 `weui-panel`/媒体列表（`weui-media-box`）；搜索用 `weui-search-bar`；选规格用 `weui-half-screen-dialog`；购物车结算用 `weui-bottom-fixed-opr` 区。**放弃左分类栏 + 右列表**（已确认）。
- **TabBar 换 weui tabbar 组件**：用 `miniprogram_npm/weui-miniprogram/tabbar` 替换自定义 `custom-tab-bar-comp`（home/invite/menu/profile 四页）。
- **token 移除放最后**：先把每页迁到 weui（不再引用 `var(--brand-*)`），全部迁完再删 token + `DESIGN.md` + 配色测试，避免中途样式塌掉。
- **保留业务逻辑与数据流**：order_type / SKU / 外卖 / 返利等 JS 逻辑不变，仅换 UI 层；JS 行为测试尽量保留。
- **不改后端**。

## Dependency Graph
```
weui 组件全局注册 (app.json)                [Phase 0, 加法, 不破坏]
   │
   ├── weui tabbar 替换 custom-tab-bar       [Phase 1, 共享 chrome]
   │
   └── 逐页迁移到 weui + 默认色 (不再引用 token)  [Phase 2, 纵切, 每页可独立验证]
          login → home → menu(重构) → order-confirm → profile → invite → share-code
                 │
                 └── 移除 DESIGN.md + app.wxss token + 配色测试 + 废弃 custom-tab-bar  [Phase 3, 收尾]
```
顺序：先注册（加法）→ 换共享 TabBar → 逐页去 token 化 → 最后删 token/DESIGN/测试（此时无人引用）。

---

## Task List

### Phase 0: weui 组件注册（加法，不破坏）

#### Task 1: 全局注册所需 weui 组件
**Description:** 在 `app.json` `usingComponents` 注册后续要用的 weui 组件（tabbar、navigation-bar、searchbar、dialog、half-screen-dialog、grids 等），保留现有 token 不动（页面仍正常）。
**Acceptance criteria:**
- [ ] `app.json` 注册 weui tabbar / searchbar / half-screen-dialog / dialog / grids（按需）。
- [ ] 编译无报错，现有页面显示不变。
**Verification:**
- [ ] 微信开发者工具编译通过；`frontend` jest 全绿。
**Dependencies:** None
**Files likely touched:** `frontend/app.json`
**Estimated scope:** S

#### Checkpoint A — 注册
- [ ] 编译干净、页面无变化、jest 全绿。

---

### Phase 1: 底部导航换 weui tabbar

#### Task 2: 用 weui tabbar 替换 custom-tab-bar（4 页）
**Description:** 用 `weui-miniprogram/tabbar` 组件替换 home/invite/menu/profile 的 `custom-tab-bar`，默认配色；tab 切换路由逻辑保持（点餐/邀请/我的）。
**Acceptance criteria:**
- [ ] 四页底部为 weui tabbar，默认绿选中态。
- [ ] 三个 Tab 切换正常，深链/扫码后落点不变。
- [ ] 不再有页面引用 `custom-tab-bar`（组件文件暂留，Phase 3 删）。
**Verification:**
- [ ] 人工：四页切 Tab 正常；jest 全绿。
**Dependencies:** Task 1
**Files likely touched:** `frontend/pages/{home,invite,menu,profile}/index.{wxml,json,js}`
**Estimated scope:** M

#### Checkpoint B — TabBar
- [ ] 四页 weui tabbar 正常，路由不回归。

---

### Phase 2: 逐页迁移到 weui + 默认配色（每页一个纵切）

> 每个任务：该页**只用 weui 组件/类 + 默认色**，移除该页 `var(--brand-*)/var(--color-*)` 自定义色引用与定制视觉；JS 行为不变。

#### Task 3: login 页 → weui
**Acceptance criteria:** [ ] 登录页用 weui（btn/cells），默认色；登录流程不变。
**Verification:** [ ] jest 全绿；人工登录走通。
**Dependencies:** Task 2 **Files:** `frontend/pages/login/index.*` **Scope:** S

#### Task 4: home 启动页 → weui
**Description:** 堂食/外卖两入口改用 weui（`weui-grids` 或 weui cells/`weui-btn`），默认色；外卖 chooseAddress 逻辑不变。
**Acceptance criteria:** [ ] 两入口为 weui 组件、默认色；堂食扫码、外卖选址逻辑不变。
**Verification:** [ ] jest（home-launcher）全绿；人工两入口可点。
**Dependencies:** Task 2 **Files:** `frontend/pages/home/index.*` **Scope:** M

#### Task 5: menu 菜单页 → 纯 weui 重构（最大）
**Description:** 重构为：顶部 `weui-search-bar` + `weui-navbar` 分类页签 + 各分类 `weui-media-box` 列表；选规格用 `weui-half-screen-dialog`；加购/数量用 weui 按钮；结算条用 `weui-bottom-fixed-opr`。移除左右两栏与自定义色。`selectCategory` 改为 navbar 页签切换；SKU/购物车/order_type/delivery 逻辑不变。
**Acceptance criteria:**
- [ ] 顶部分类页签（weui-navbar）切换对应分类列表；无左右两栏。
- [ ] 有规格→half-screen-dialog 选规格；无规格→weui 加购/数量。
- [ ] 购物车结算条为 weui；堂食/外卖、SKU 计价、去结算逻辑不变。
**Verification:**
- [ ] 重写后的 `menu-page.test.js`（页签切换 + switchOrderType）全绿；cart-sku 测试全绿；人工分类切换/选规格/加购/结算走通。
**Dependencies:** Task 2 **Files:** `frontend/pages/menu/index.*`, `frontend/__tests__/menu-page.test.js` **Scope:** L

#### Task 6: order-confirm → weui
**Description:** 用 weui cells/媒体列表与 `weui-bottom-fixed-opr` 重排；移除 `order-type-tag`/`delivery-addr` 等自定义色样式，改 weui；逻辑（金额/福利金/order_type/地址）不变。
**Acceptance criteria:** [ ] 全 weui、默认色；下单/支付/外卖地址展示逻辑不变。
**Verification:** [ ] order-confirm 相关测试全绿；人工下单流程走通。
**Dependencies:** Task 2 **Files:** `frontend/pages/order-confirm/index.*` **Scope:** M

#### Task 7: profile 我的 → weui
**Description:** 去掉渐变头部/钱包卡定制色，改 weui cells/panel + 默认色；收支明细/订单列表用 weui 列表。
**Acceptance criteria:** [ ] 全 weui、默认色；钱包/订单/明细展示不变。
**Verification:** [ ] jest 全绿；人工查看我的页。
**Dependencies:** Task 2 **Files:** `frontend/pages/profile/index.*` **Scope:** M

#### Task 8: invite 邀请 → weui
**Acceptance criteria:** [ ] 去渐变头部，改 weui + 默认色；邀请码/分享逻辑不变。
**Verification:** [ ] jest（share-button 等）全绿；人工邀请页走通。
**Dependencies:** Task 2 **Files:** `frontend/pages/invite/index.*` **Scope:** M

#### Task 9: share-code 分享码 → weui
**Acceptance criteria:** [ ] 去渐变头部，改 weui + 默认色。
**Verification:** [ ] jest 全绿；人工查看分享码页。
**Dependencies:** Task 2 **Files:** `frontend/pages/share-code/index.*` **Scope:** S

#### Checkpoint C — 全页迁移
- [ ] 全部页面 weui + 默认色；`grep -r 'var(--brand' frontend/pages` 无命中；jest 全绿；人工各页观感为 weui 原生。

---

### Phase 3: 移除定制体系（收尾）

#### Task 10: 删 DESIGN.md + app.wxss token + 配色测试 + 废弃组件
**Description:** `app.wxss` 移除 `--brand-*`/`--color-*`/`--weui-*` 覆盖与定制工具类，仅保留 weui 导入与必要全局（布局类如 `.page`）；删除 `DESIGN.md`；删除 `frontend/__tests__/design-tokens.test.js`（其断言的定制 token 已不存在）；删除已无引用的 `custom-tab-bar-comp`（及废弃 `custom-tab-bar/tab-bar/`）；清理 CLAUDE/SPEC 中对 DESIGN.md 的引用（如有）。
**Acceptance criteria:**
- [ ] `app.wxss` 无 `--brand-*`/品牌色覆盖；`DESIGN.md` 已删。
- [ ] 全仓 `grep -rn 'DESIGN.md'` 无残留引用（除归档计划）。
- [ ] `grep -rE 'var\(--(brand|color)' frontend`（排除 weui 库）无命中。
- [ ] 删除配色测试；其余 jest 全绿。
**Verification:**
- [ ] `frontend` jest 全绿；编译通过；人工全站为 weui 默认绿。
**Dependencies:** Checkpoint C
**Files likely touched:** `frontend/app.wxss`, `DESIGN.md`(删), `frontend/__tests__/design-tokens.test.js`(删), `frontend/miniprogram_npm/custom-tab-bar-comp/*`(删), `frontend/app.json`
**Estimated scope:** M

#### Checkpoint D — 完成
- [ ] 无 DESIGN.md、无定制 token、无 custom-tab-bar；全站 weui 原生 + 默认绿；jest 全绿；人工验收。

---

## Risks and Mitigations
| Risk | Impact | Mitigation |
|------|--------|------------|
| 提前删 token 导致样式塌 | High | Token 删除（T10）放最后；Checkpoint C 用 grep 确认无引用再删 |
| 菜单重构改变交互/破坏 selectCategory 测试 | Med | T5 同步重写 `menu-page.test.js`（navbar 页签）；保留 SKU/cart 逻辑测试 |
| weui tabbar 组件 API 与自定义不同（可能需 list 配置/选中态联动） | Med | T2 先查 weui tabbar 组件用法（context7 / 组件源码）再接；逐页验证路由 |
| 大量自定义 WXSS 删除遗漏，残留冷暖混搭 | Med | 每页迁移后 grep 该页无 `var(--brand`；Checkpoint C 全量 grep |
| 渐变头部（profile/invite/share-code）无 weui 等价 | Low | 用 weui 默认页头/cells 平铺；接受视觉简化（已确认去定制） |
| 默认绿与此前橙色品牌差异大 | Low | 已确认回到 weui 默认 |

## Open Questions
1. **菜单顶部搜索**：是否需要 `weui-search-bar`（图中有搜索框），还是本期仅分类页签 + 列表？（计划含 searchbar，可删）
2. **profile/invite 渐变头部**：接受改为 weui 纯色/cells 的视觉简化？（计划默认接受）
3. **深色模式**：本期仅浅色（已确认）；是否预留 `data-weui-theme` 钩子？（计划不预留）

## Not Doing（及原因）
- **深色模式** —— 已确认仅浅色。
- **改后端 / 业务逻辑** —— 仅 UI 层去定制化。
- **管理后台(Vue)** —— 需求针对小程序。
- **保留任何品牌定制色** —— 与目标相反。
