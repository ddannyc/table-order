# Plan：实现 v6 设计稿（首页 / 菜品 / 下单 / Tab）

**视觉真源**：`docs/design/mockup-v6.png`（+ 可在浏览器打开 `mockup-v6.html`）。已定稿。
目标：把真实小程序三屏 + tab 改到与 v6 一致。沿用上轮：纵切、一任务一提交、TDD。
**基线**：绝不 git add `frontend/config.js` / `.claude/` / `specs/`；每次只 stage 该任务文件，禁止 `git add -A`。

## 诚实说明（关键）
本环境**无 WeChat 模拟器**，无法像 comp 那样渲染真实 wxml 截图。所以：
- 自动化只能测**结构断言 + 对比度 + 无回归**。
- **像素级对齐 v6 必须你在真机/开发者工具里核对**（C1）。comp 是规格，不是测试替身。
- comp 里的 SVG（厨师 / 菜品插画 / 分段图标）将原样移植成 wxss `background-image` data-URI 或内联，颜色与令牌同源。

## 与现有代码的差异（R1–R5 已实现，但不匹配 v6）
- **首页**：现为"绿 band + 余额/返利 + 厨师 + 两入口卡"；v6 要"品牌头(无余额/福利金) + 堂食/外卖分段(线描图标) + 扫码点餐卡 + 厨师 banner(v1版) + 福利放送"。**需移除 R2 的余额/返利**。
- **菜品**：现占位是陶土橙字形块；v6 要**插画菜品图**（碗/盘/煲 SVG）。购物车条需压在 tab 之上。
- **下单**：现为福利金 toggle + amount card；v6 要"福利金抵扣开关 + 支付方式仅微信 + 菜品缩略图 + 店头行"。
- **Tab**：现为重染线描 PNG；v6 tab 线描风格，需对齐。

---

## 配色令牌（对齐 comp）
```
--weui-BRAND #234B3A（comp 用色，比现 #2C4A3B 略深）  --green-2 #2F6B4F
--weui-BG-0 #F3EEE4  --weui-BG-1 #FFFFFF/#FBF8F2  --accent #C8643C  --price-ink #B0491F
```

## 依赖图
```
D1 令牌对齐 ──┬─→ D2 首页重构到 v6
              ├─→ D3 菜品页到 v6（插画菜品图 + 购物车条层级）
              │      └─→ D4 下单页到 v6（福利金开关 + 仅微信 + 缩略图）
              └─→ D5 Tab 线描图标对齐
                          └─→ C1 真机核对 v6 一致性
```
顺序：D1 → D2 → D3 → D4 → D5 → C1。

---

## D1 — 配色令牌对齐 v6 — S
**改** `app.wxss`：`--weui-BRAND` → `#234B3A`，新增 `--green-2:#2F6B4F`；其余令牌不变。更新 `theme-tokens.test.js` 期望值 + 对比度（墨绿更深，对比只增不减）。
**验收**：令牌更新；对比度测试过；全量 jest 绿。

## D2 — 首页重构到 v6 — L（依赖 D1）
**改** `pages/home/index.{js,wxml,wxss}`。
- **移除** R2 的余额/返利（含 `loadWallet`、品牌头里的 balance/reward 绑定与 wallet API import）。
- 品牌头：logo 字标 + "四月春膳" + "一汤一蔬·阅沐春风"（**无**余额/福利金）。
- 分段 **堂食/外卖**（线描 cloche / scooter 图标，currentColor）。交互：选「堂食」→ 扫码点餐卡触发 `scanDineIn`；选「外卖」→ 触发 `chooseDelivery`（resolveDeliveryShop→menu）。
- 扫码点餐卡（→ `scanDineIn`）。
- 厨师 banner（移植 **v1 版**厨师 SVG，data-URI 背景）。
- **福利放送**：见 Open Q（数据源待定）。
**验收**：`home-launcher.test.js` 行为（堂食扫码/外卖解析/兜底）全绿；移除 wallet 后无残留引用（grep 干净）；结构断言（品牌头无 balance、分段含线描图标、banner data-URI、扫码卡）。
**验证**：全量 jest 绿。

## D3 — 菜品页到 v6 — L（依赖 D1）
**改** `pages/menu/index.{wxml,wxss}`、`utils/menu-image.js`。
- 占位从"陶土橙字形块"升级为 **comp 的插画菜品图**（碗/盘/煲/丸 SVG，按分类映射），data-URI 背景，暖色 tile 底。`product.image` 有则用真图，空走插画。
- 购物车结算条 `bottom` 对齐 tab 高度（**不重叠**，沿用 comp 修复）；list 底部留白清开 cartbar+tab。
- 圆形加购、类目数量徽章（已存在，保留）。
**验收**：`menu-image.test.js`/`menu-reskin.test.js`/`menu-page.test.js` 全绿；结构断言（插画 data-URI、cartbar bottom 不为 0/不压 tab）。
**验证**：全量 jest 绿。

## D4 — 下单页到 v6 — M（依赖 D3）
**改** `pages/order-confirm/index.{wxml,wxss}`（逻辑不破坏）。
- 店头卡：店名 + 距离 + 电话/桌号行。
- 商品明细：菜品缩略图（`item.image` 空走插画）。
- **福利金抵扣开关**（switch，沿用既有 useReward 逻辑）+ 明细里 `-¥xx` 优惠行。
- **支付方式：仅微信支付**（绿微信图标 + 选中勾；去掉"福利金当支付方式"）。
**验收**：既有 `order-confirm-*`/`api-create-order` 行为测试全绿（**无支付逻辑回归**）；结构断言（仅一种支付方式、福利金为 switch 非 radio、缩略图）。
**验证**：全量 jest 绿。

## D5 — Tab 线描图标对齐 v6 — S/M（依赖 D1）
**改** `static/*.png`（必要时重生成）、`utils/tabbar.js`。
- 让 tab 图标与 v6 线描风格一致（点餐=店铺/碗线描、我的=人物线描），墨绿激活/灰未激活。我用脚本生成 PNG → **用图片读取器自检**；不达标则请你提供。
**验收**：`tab-icons.test.js` 断言激活图标为墨绿；引用路径有效。
**验证**：全量 jest 绿。

---

## Checkpoint C1 — 真机核对 v6 一致性
- 全量 jest 绿（记录数字）。
- **人工（必须）**：开发者工具/真机逐屏对照 `mockup-v6.png` —— 首页/菜品/下单/tab 是否一致；data-URI SVG 是否渲染；插画/图标观感。
- 不一致项回流为下一轮微调。

## Resolved Decisions（已定）
1. **福利放送 = 静态装饰卡**：不接数据，两张插画装饰卡（comp 观感），内容为占位、不可点；标注"静态占位、日后可接精选菜品"。不是假数据模块（无后端调用、不声称真实）。
2. **堂食/外卖分段交互 = 外卖即进配送流**：选「外卖」→ `chooseDelivery`（resolveDeliveryShop→menu）；扫码点餐卡 = `scanDineIn`（仅堂食）。

## 不在本期
- 真实菜品摄影（OSS 后填，插画过渡）；优惠券/自提/预约时间（无后端）；地址/选店/我的订单屏。
