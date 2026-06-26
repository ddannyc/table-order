# Plan：菜单左右布局 + 入口即定单型（含前端设计规范）

来源：`/idea-refine` 一页纸 + 用户决策。范围：纯前端（小程序）改版，后端不变。
基线：绝不 `git add` `frontend/config.js` / `.claude/` / `specs/`；一任务一提交，仅 stage 该任务文件 + todo；TDD（RED→GREEN→回归 jest→commit）。

## 已确认决策（来自 /idea-refine）
- 图片：**纯 CSS/SVG 占位图**（不引入二进制资源），按分类兜底（seed 不塞图；`product.image` 为空时前端回退到分类 CSS/SVG 占位块）。
- 切换：**移除菜单内堂食/外卖切换**；菜单只读显示模式；**堂食与外卖购物车互不共享**；改模式需回首页重新进入。
- 右侧菜品行：**大图卡片（美团/瑞幸风）**，图片为一等公民。
- 单型在**首页唯一入口**确定：扫码绑桌→堂食；首页外卖卡→外卖。

## 设计方向（遵循 /frontend-design，钉死在既有 weui token 体系内）
既有视觉方向已"钉死"：weui 组件 + `app.wxss` 的间距/字号比例 + 品牌色 `--weui-BRAND`（weui 绿）。设计技能在此的职责是**在既有体系内做克制而有意的版式**，不另起配色。

- **Color**：背景 `--weui-BG-1`；卡片 `#fff`；主文本 `--weui-FG-0`；次要 `--weui-FG-2`；强调（激活分类指示条 / 价格 / 加购按钮）仅用 `--weui-BRAND`。不引入新色。
- **Type**：沿用既有比例。菜品名 `--font-subhead` 600；描述 `--font-caption` `--weui-FG-2`；价格 `--font-price` 600 `--weui-FG-0`，"起"用 `--font-caption`/`--weui-FG-2`。类目名 `--font-body`。
- **Layout（左右布局，过滤式，不做 scroll-spy）**：左侧竖向类目轨（固定宽 ~180rpx，吸顶、可滚），右侧当前类目的大图卡片列表（可滚）。点左轨 → 右列表切到该类目（沿用现有 `activeCategory` + `productsByCategory` 过滤，零新增数据模型）。
  ```
  ┌──────── menu-shopbar（只读模式标识：堂食·桌号A01 / 外卖配送）────────┐
  ├────────┬──────────────────────────────────────────────────────────┤
  │ 新品   │  ┌──────┐  招牌奶茶                                        │
  │[新品 ▌]│  │ img  │  经典招牌奶茶                                     │
  │ 芝士   │  └──────┘  ¥13起                       [选规格] / [− 1 +]   │
  │ 奶茶   │  ┌──────┐  厚乳波波                                        │
  │ 气泡水 │  │ img  │  厚乳 + 黑糖波波              ¥14        [加入]    │
  └────────┴──────────────────────────────────────────────────────────┘
            （底部购物车结算条 + tabbar 沿用现有）
  ```
- **Signature**：左轨**激活态**——激活类目用 weui 绿竖条指示 + 白底凸出（其余类目灰底次要色）。这是页面唯一"亮点"，其余保持安静。
- **a11y/质量地板**：触控目标 ≥ 80rpx；激活态不只靠颜色（加竖条/字重）；尊重 `prefers-reduced-motion`（类目切换/加购动效可禁用）；占位图带 `aria`/alt 语义（mode="aspectFill" + 背景兜底，避免裂图）。
- **Copy（界面口吻）**：只读模式标签——堂食 `堂食 · 桌号{X}`、外卖 `外卖 · 送货上门`；空菜单 `该店铺还没有上架菜品`；图片加载失败回退占位图（不显示裂图、不弹错）。动作词与现状一致（加入/选规格/去结算）。

## 依赖图（纵切，一任务一条完整路径）
```
M1 入口即定单型(移除切换+只读模式+单型可靠推导)
        │（提供可靠 orderType）
        ▼
M2 购物车按单型隔离(storage/product/menu/order-confirm)
M3 本地占位图 + 图片兜底  ──┐（M3 独立，可先于 M4）
        │                   ▼
        └────────────►  M4 左右布局 + 大图卡片(用 M3 的图、M1 去掉的切换)
        ▼
M5 Checkpoint：设计自审 + 全量 jest 回归
```

## 任务

### M1 — 入口即定单型：移除菜单切换 + 只读模式 + 单型可靠推导 — M
**做什么**
- 删除菜单内 `switchOrderType` 切换按钮（wxml/js）及相关 wxss；顶部改为**只读模式标识**（堂食 `堂食 · 桌号{boundTableNo}` / 外卖 `外卖 · 送货上门`）。
- 单型推导收敛：进入菜单时 **delivery 标识优先 → 外卖；否则有桌号绑定 → 堂食**；外卖进入时清空桌号绑定（避免历史堂食绑定泄漏）；把当前 orderType 持久化随绑定存储，使 reLaunch 重进可正确复原。
**验收**
- 扫码进入：标识显示 `堂食 · 桌号A01`，无外卖态。
- 首页外卖卡进入：标识显示 `外卖 · 送货上门`，无桌号、无残留堂食绑定。
- 菜单→购物车→返回，模式不变；菜单内无任何"切换"控件。
**验证/测试**（`__tests__/menu-page.test.js` 等）
- 断言 `switchOrderType` 已移除/不再被绑定；只读标识按模式渲染。
- 推导：delivery 入参→delivery；仅桌号绑定→dine_in；外卖入参清桌号。
- `node node_modules/jest/bin/jest.js menu-page`

### M2 — 购物车按单型隔离 — M（依赖 M1）
**做什么**
- 购物车存储键由 `shopId` 扩为 `shopId + orderType`，贯穿 `utils/storage.js`、`frontend/api/product.js`（getCart/setCart/clearCart/getCartTotal/getCartCount/addToCart/updateCartQuantity 增 orderType 维度）、`pages/menu/index.js`（所有购物车调用传 orderType）、`pages/order-confirm/index.js`（按 shopId+orderType 读取结算）。
- 兼容：旧单一 shopId 键的处理（迁移或忽略，dev 期可直接新键，不做迁移——seed/测试环境无存量）。
**验收**
- 先建堂食购物车 → 回首页进外卖 → 建外卖购物车：两份互不影响；各自结算行项正确（堂食含桌号、外卖含配送字段）。
- 清空一种模式的购物车不影响另一种。
**验证/测试**（新增 `__tests__/cart-isolation.test.js` 或扩 `storage`/`product` 测试）
- 同 shop 不同 orderType 的 add/get/clear 相互独立。
- order-confirm 读取与当前模式一致的购物车。
- `node node_modules/jest/bin/jest.js`

### M3 — CSS/SVG 分类占位图 + 图片兜底 — S（独立）
**做什么**（决策：**不引入二进制资源**，用纯 CSS/SVG 画占位图）
- 占位图用**内联 SVG / 纯 CSS**：按分类的字形（如奶茶杯/芝士/气泡/新品）置于**中性浅底块**（`--weui-BG-2`/`--weui-FG-3` 底 + `--weui-FG-2` 字形）——不引入新配色，把"亮点"留给左轨。
- 助手 `placeholderFor(category)` → `{ glyph, ... }`（含通用兜底）；卡片缩略图：`product.image` 存在 → `<image>`；否则渲染占位块（CSS/SVG）。真实 `<image>` `binderror` → 切到同款占位块（不裂图、不弹错）。
**验收**
- seed 无图菜品：渲染该分类的 CSS/SVG 占位块。
- 有 `image` 的菜品：用真实图；加载失败回退占位块。
- 未知分类：通用兜底字形。
**验证/测试**（新增 `__tests__/menu-image.test.js`）
- `placeholderFor`：已知分类→对应字形；未知→通用兜底。
- 卡片逻辑：有 image→走 `<image>`；无 image→走占位块。
- `node node_modules/jest/bin/jest.js menu-image`

### M4 — 左右布局 + 大图卡片 — L（依赖 M1、M3）
**做什么**
- `menu/index.wxml`：左侧竖向类目轨（`scroll-view scroll-y`，吸顶，激活态绿竖条+白底）+ 右侧大图卡片 `scroll-view`；点轨切 `activeCategory`（沿用过滤）。
- 大图卡片行：左缩略图（`resolveProductImage`，`aspectFill`，圆角 `--radius-md`）+ 右名称/描述/价格/加购；售罄、规格选择、`+/−` stepper 沿用现有逻辑与节点。
- `menu/index.wxss`：左右栅格、卡片、激活指示，全部用既有 token；落实 a11y 地板（触控≥80rpx、激活非仅靠色、`prefers-reduced-motion`）；注意选择器特异性，避免 section/element 规则相互抵消的 padding/margin 问题。
**验收**
- 模拟器：左轨点选切换右列表；卡片显示图片（占位/真实）；加购/步进/规格/售罄均正常；小屏不溢出、购物车条与 tabbar 不遮挡。
- 视觉：仅激活类目与价格/加购用绿，其余安静；无裂图。
**验证/测试**
- `selectCategory` 与卡片渲染逻辑测试；更新既有 `menu-page` 断言至新结构（不再有横向 navbar/切换）。
- `node node_modules/jest/bin/jest.js`

### M5 — Checkpoint：设计自审 + 回归 — S
- 设计自审（/frontend-design）：signature 是否唯一且克制？是否可"摘掉一个配饰"？copy 是否界面口吻、动作词一致？a11y 地板达标？
- 全量三态绿：`node node_modules/jest/bin/jest.js`（目标 ≥ 现有 95）；后端不动无需重跑 go/vitest（除非触及，预期不触及）。
- 模拟器截图核对左右布局与大图卡片。

## Checkpoint 汇总
- [ ] M1 菜单无切换、只读模式标识、单型推导可靠
- [ ] M2 堂食/外卖购物车隔离，各自结算正确
- [ ] M3 无图菜品显示分类 CSS/SVG 占位块，失败兜底不裂图
- [ ] M4 左右布局 + 大图卡片，交互/售罄/规格沿用，token 合规、a11y 达标
- [ ] M5 设计自审通过；全量 jest 绿

## 不在本期
- 真实菜品摄影 / 后台图片上传 UI（本期本地占位图，前端兜底已为真实图留口）
- 类目 scroll-spy / 锚点联动（沿用过滤式，更简单）
- 就近门店、菜单搜索
- 菜单内任何"切换点餐方式"入口（决策：首页唯一入口）

## 待澄清（可在 M1/M4 边做边定，非阻断）
- 只读模式标识落位：保留 `menu-shopbar` 内，还是并入左轨头部省竖向空间？（倾向 shopbar 内，改动最小）
- 占位图：每分类一张 vs 一张通用？（倾向每分类一张，避免"一墙相同卡片"）
