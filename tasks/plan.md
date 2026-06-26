# Plan：复刻 mp-ui（首页 / 菜品 / 下单 / Tab）

来源 spec：`docs/ideas/replicate-mp-ui.md`（取代上一轮 pine-ink 方向）。
风格沿用：纵切、一任务一提交、TDD（RED→GREEN→回归 jest）。
**基线**：绝不 git add `frontend/config.js` / `.claude/` / `specs/`；每次只 stage 该任务文件，禁止 `git add -A`。

---

## 关于"测试"的诚实说明（同上轮）
可自动化 = ①WCAG 对比度（jest 内算）②结构断言（读 wxml/wxss/js）③行为无回归（既有 jest 全绿）。
**不可自动化** = 复刻保真度、彩色插画好不好看、PNG 图标质量、真机渲染 —— C1 人工核对。

## 两个保真天花板（提前声明，避免又一次失望）
- **彩色厨师插画**用矢量重绘，是**近似**参考图意象，非像素级一致。
- **Tab PNG 图标**在本环境生成质量有限；若我生成的不够好，需你提供一套（已在 spec Open Q）。

## 配色令牌（最终值）
```
--weui-BRAND #2C4A3B 墨绿   --weui-BG-0 #F3EEE4 米白   --weui-BG-1 #FBF8F2 卡面
--weui-FG-0 #2A2723         --weui-FG-2 #6E665A         --weui-FG-3 #E3DCCE
--accent     #C8643C  陶土橙：标签/徽章/加购点缀（填充；文字慎用，小字对比不足）
--price-ink  #B0491F  深陶土：价格（4.76:1 ✓，正文级也过）
删除 --brand-accent（金退场）。
```
真数据来源：余额 `getBalance()`→`{balance,reward_balance}`；返利 `getRewardBalance()`。

---

## 依赖图
```
R1 令牌(金→陶土橙) ──┬─→ R2 首页融合 dashboard
                     ├─→ R3 菜品照片卡 + 类目徽章 ──→ R4 下单复刻（共享购物车带图）
                     └─→ R5 底部 tab 墨绿 + PNG 图标
                                        └─→ C1 自审 + 全量绿 + 真机核对
```
顺序：R1 → R2 → R3 → R4 → R5 → C1。

---

## R1 — 配色令牌改造（金→陶土橙）— S
**改**：`frontend/app.wxss`（删 `--brand-accent`，加 `--accent:#C8643C`，`--price-ink`→`#B0491F`）；`__tests__/theme-tokens.test.js`（断言新值 + 对比度：price-ink on BG-0/BG-1 ≥4.5；accent 仅作填充不设文字门槛）。
**验收**：令牌更新；对比度测试过；既有 jest 全绿。
（注：home hero / menu glyph 仍内联金 SVG，将在 R2/R3 被彩色插画/照片卡替换——R1 后短暂并存可接受。）
**验证**：`cd frontend && node node_modules/jest/bin/jest.js` 全绿。

## R2 — 首页融合 dashboard — L（依赖 R1）
**改**：`pages/home/index.js/.wxml/.wxss`。
- 墨绿品牌头：店名/品牌字标 + **余额**（`getBalance`）+ **返利**（`getRewardBalance`），未登录则显 `—` 并引导登录，**不假数据**。
- **彩色厨师插画 banner**（矢量重绘，替换 T3 单线 hero）。
- 堂食/外卖入口卡（参考图卡片样式：图标/标题/描述/箭头）。
**验收**
- 余额/返利走真接口；未登录优雅降级（结构测试 + js 行为：fetch 调用、降级分支）。
- `home-launcher.test.js`（堂食扫码/外卖解析/兜底）**全绿**；新增结构断言（品牌头/banner/入口卡）。
- 无硬编码假余额。
**验证**：jest 全绿。

## R3 — 菜品照片优先卡 + 类目数量徽章 — L（依赖 R1）
**改**：`pages/menu/index.js/.wxml/.wxss`、`utils/menu-image.js`、`api/product.js`/`utils/storage.js`（购物车存 `image`，供 R4 下单缩略图）。
- 照片优先卡：大 thumb 用 `product.image`，空则**设计过的彩色展位图**（非空块）；陶土橙价格；绿色圆形加购按钮。
- 左类目轨加**数量徽章**（真数据 = 该类目菜品数）。
- 加购时把 `image` 写进购物车项。
**验收**
- 卡片读 `product.image`、空走展位图（`menu-image.test.js` 扩展：resolveProductImage 返回展位图描述）。
- 类目徽章数 = 菜品数（js 测试）。
- 购物车项含 `image`（`cart-isolation.test.js` 或新测试断言）。
- `menu-page.test.js` 全绿。
**验证**：jest 全绿。

## R4 — 下单复刻 — M（依赖 R3）
**改**：`pages/order-confirm/index.js/.wxml/.wxss`。
- 商品明细加**照片缩略图**（读购物车项 `image`，空走展位图）。
- 店头 + 支付方式（**余额支付** real：`balance` 充足才可选 + 微信支付）+ 合计/支付 + **支付成功态**（参考图勾选浮层）。
- 复刻视觉，逻辑（福利金抵扣等既有）不破坏。
**验收**
- 明细渲染缩略图（结构断言）；支付方式含余额/微信（结构）。
- 既有下单/福利金/配送费逻辑测试**全绿**（无行为回归）。
- 支付成功态可触发（js/结构）。
**验证**：jest 全绿。

## R5 — 底部 tab 墨绿复刻 + PNG 图标 — M（依赖 R1）
**改**：`utils/tabbar.js`（图标路径）、新增 `/static/*.png` 墨绿图标、相关 tab wxss。
- 生成一套扁平墨绿 tab 图标（normal/active）；mp-tabbar 复刻参考图配色。
- **若生成质量不足 → 停下请你提供素材**（doubt-driven，资产类不硬撑）。
**验收**
- tabbar 引用新图标；三页 tab 行为不回归（既有 tab 测试/手测）。
- 新增断言：tabbar.js 指向新图标路径。
**验证**：jest 全绿。

---

## Checkpoint C1 — 设计自审 + 全量绿 + 真机核对
- /frontend-design 自审（保真优先；signature=厨师 banner；a11y：对比度实测、触控≥80rpx、激活非纯色；克制）。
- 全量 jest 绿（记录数字）。
- **人工待办**：真机核对 ① 展位图够不够体面 ② 彩色插画/PNG 图标观感 ③ 余额/返利真机取数 ④ 整体是否达参考图保真级别。达标后再决定铺其余屏。

## 不在本期
- 优惠券 / 自提 / 预约时间（无后端）；真实菜品摄影（OSS 后填）；菜单内 自提/外卖 切换（M1 已删）；地址/选店/我的订单屏。
