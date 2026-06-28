# 实施计划：修复菜单卡片「选规格」按钮在三位数价格下换行（方案 V3）

## 问题诊断（事实来源）
菜单卡片底部 `.menu-card-bottom`（`flex; justify-content:space-between`）放着**黄底大价格标** `.menu-price` 和 `选规格` 按钮。价格恒 `toFixed(2)`，三位数 → `¥100.00起`（"100.00" 6 字符）把价格标撑宽；真机卡片正文仅 ~145px（左轨 90 + 缩略图 80 吃掉大半），价格 + 默认 `weui-btn_mini` 按钮放不下一行，按钮被压窄 → `选规格` 三字**断成两行**。

## 已定方案（4 套效果图、375px 真机比例验证后）
**V3 — 瘦黄标 + 整数去 .00 + 紧凑按钮**：保留大食物图与黄色价格身份；把价格标做小一档、整数价不显小数（`¥100`，非整价如 `¥38.50` 仍保留小数），并把 `选规格` 做成紧凑原子药丸。实测此组合下 `¥38 / ¥100 / ¥288 / ¥1288 起` **全部价格档价格 + 按钮同一行**。保留 `flex-wrap` 仅作极端兜底（永不破词）。

效果图对照：V1（大黄标+.00，按钮总掉次行，已否决）/ V2（去黄框纯文字，仍换行，否决）/ V3（选定）/ V4（缩小缩略图，备选）。

## 现状（已确认，勿改逻辑）
- 价格来源：`menu/index.js` `loadData` 里 `priceText = p.price.toFixed(2)`、`specMinText = Math.min(...specs.map(s=>s.price)).toFixed(2)`；卡片显示 `p.hasSpecs ? p.specMinText : p.priceText`。规格弹层等其它价格**本期不动**。
- `.menu-action` 三分支：`售罄` / `选规格`（唯一 weui-btn，会断行）/ `+`·加减器（定宽，不受影响）。
- 价格标 `.menu-price`：黄底红字 + `--font-price(34rpx)` + 立体影。V3 仅**瘦身**，不改色彩身份。

## 架构决策
1. **价格格式化抽成纯函数** `formatPrice(n)`：整数→无小数（`100`），非整数→保留两位（`38.50`）。供菜单卡片价格用，Vitest/jest 守护。仅作用于**菜单卡片**的 `priceText`/`specMinText`，不波及购物车/下单页（各自语境保留精度）。
2. **价格标瘦身**：缩小 `.menu-price` 字号/内边距（仍黄底红字，去重立体影或减弱），降低占宽。
3. **按钮紧凑原子化**：`选规格` 加专属类 `menu-spec-btn`（保留 `weui-btn weui-btn_primary` 粉底白字，接管尺寸）：更小字号 + 紧 padding + `white-space:nowrap` + `flex-shrink:0`；`.menu-action { flex:none; margin-left:auto }`。
4. **兜底不破词**：`.menu-card-bottom { flex-wrap:wrap; row-gap:var(--space-xs) }`，万一某极端组合仍超宽，整颗按钮掉次行右对齐，绝不把 `选规格` 拆字。
5. **可访问性**：按钮高度 ≥ ~56rpx 触控目标；纯展示层改动。
6. **测试**：纯函数 `formatPrice` 单测守护；UI 用源码字符串守护类名/样式键；视觉适配真机多档价格手测。

## 依赖图
```
T1 价格格式化纯函数 formatPrice（utils/price.js）+ 接入 menu loadData
        │
        ▼
T2 价格标瘦身 + 选规格 紧凑原子按钮 + 行兜底（menu wxml 加类 + wxss 样式）
```

## 任务（垂直切片）

### T1 — 价格格式化纯函数 + 接入菜单  〔S〕
**描述**：新增 `frontend/utils/price.js` 的 `formatPrice(n)`：整数去掉 `.00`，非整数保留两位；在 `menu/index.js` `loadData` 用它生成卡片的 `priceText`/`specMinText`。
- `formatPrice(100)→"100"`、`formatPrice(38.5)→"38.50"`、`formatPrice(38)→"38"`、`formatPrice(0)→"0"`、容错字符串数字。
- `menu/index.js`：`priceText: formatPrice(p.price)`、`specMinText: formatPrice(Math.min(...specs.map(s=>s.price)))`。规格行内 `s.price.toFixed(2)` 等**不动**（本期仅卡片主价）。
**验收**：
- [ ] 整数价显示无小数（`¥100起` / `¥38起`），非整价保留两位（`¥38.50`）。
- [ ] 既有菜单测试不回归；新增 `formatPrice` 单测全绿。
**验证**：`cd frontend && npm test`。**文件**：`frontend/utils/price.js`、`frontend/__tests__/price-format.test.js`、`frontend/pages/menu/index.js`。

### T2 — 价格标瘦身 + `选规格` 紧凑原子化 + 行兜底  〔S〕
**描述**：缩小价格标、把 `选规格` 做成紧凑不换行药丸、操作区不收缩靠右、底行可兜底换行。
- `menu/index.wxml`：`选规格` 按钮 `class` 加 `menu-spec-btn`（移除 `weui-btn_mini`，尺寸由新类接管；保留 `weui-btn weui-btn_primary`）。
- `menu/index.wxss`：
  - `.menu-price` 瘦身（字号降一档、padding 收紧、立体影减弱/去除）。
  - `.menu-spec-btn { white-space:nowrap; flex-shrink:0; font-size:小; padding:紧; min-height≈56rpx; }`。
  - `.menu-action { flex:none; margin-left:auto; }`。
  - `.menu-card-bottom { flex-wrap:wrap; row-gap:var(--space-xs); }`。
**验收**（核心：`选规格` **永远单行不断字**）：
- [ ] `¥38 / ¥100 / ¥288 / ¥1288 起`：价格与 `选规格` **同一行**，按钮单行不断字。
- [ ] 叠加菜量徽章时仍同行不断字；极端超宽时整颗按钮掉次行右对齐（兜底）。
- [ ] 非 spec 菜（`+`/加减器）、`售罄`、购物车条不回归；按钮可点。
**验证**：`cd frontend && npm test`（守护测试全绿）；真机/DevTools 多档价格手测（重点 `¥100起`）。**文件**：`frontend/pages/menu/index.wxml`、`frontend/pages/menu/index.wxss`、`frontend/__tests__/menu-price-btn.test.js`。

### Checkpoint A（收尾）
- [ ] `cd frontend && npm test` 全绿。
- [ ] 真机手测：两位数 / 三位数（`¥100起`）/ 四位数 / 非整价（`¥38.50`）/ 带徽章 —— `选规格` 均同行不断字。
- [ ] git diff 仅含 `frontend/`（utils + menu + tests）；不碰 JS 业务逻辑/接口/`config.js`。
- [ ] 人工 review 后再决定合并 + 上传体验版（**停等用户**）。

## 守护测试映射
- T1 → `__tests__/price-format.test.js`（`formatPrice` 整数/非整数/0/容错）；既有 `menu-page.test.js` 守护 loadData 不回归。
- T2 → `__tests__/menu-price-btn.test.js`（WXML 含 `menu-spec-btn`；WXSS `.menu-spec-btn` 含 `white-space:nowrap`+`flex-shrink:0`；`.menu-card-bottom` 含 `flex-wrap:wrap`）。

## 不在本期
- 购物车/下单页/规格弹层的价格格式统一（保留各自 `.00`，如需统一再单列）。
- 缩小缩略图（V4 备选）、改价格标色彩身份、引入组件测试栈。

## 风险
| 风险 | 级别 | 缓解 |
|---|---|---|
| 菜单显 `¥100`、下单页显 `¥100.00` 观感不一致 | 低 | 下单页保精度合理；如要统一另起任务 |
| 紧凑按钮触控目标偏小 | 低 | min-height ≈56rpx + 横向 padding 给足点击区 |
| `formatPrice` 浮点边界（如 38.1）| 低 | 单测覆盖整数/一位/两位小数与字符串入参 |

## 开放问题（已关闭）
- 方案 → **V3**（瘦黄标 + 整数去 .00 + 紧凑按钮），效果图 375px 真机比例验证全档同行。
