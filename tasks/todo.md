# Todo：鸡福旺 配色落地（Pine-Ink → JFW Reskin）

详见 `tasks/plan.md`。视觉真源：`spec-shot.png`（菜单）/`screens-shot.png`（首页/结算/我的）。
TDD、一任务一提交；改代码同提交内同步它守护的测试。设备核对用 weapp-dev MCP 截图对照效果图。
**git 卫生**：只 stage 该任务文件 + `tasks/`；禁止 `git add -A`；绝不 stage `frontend/config.js`/`.claude/`/`specs/`。

## ✅ 已定（2026-06-28）
- [x] Q1 WCAG = **微调色值达 4.5:1**（不放宽测试；粉双值 + 价签红字 on 卡 + tag-red 调深）
- [x] Q2 品牌文案 = **一并改成鸡福旺**（随 T4）
- [x] Q3 页面背景 = **淡粉 `#FBEFF3`**（`--weui-BG-0`）
- [x] Q4 分支 = **当前 `feat/shansong-delivery`**

nudged 工作值：`--price-ink #D11414`(红字 on 卡)、`--pink-deep #D81A60`(小字浅色带)、`--jf-tag-red #C2185B`(优惠圈字)；亮黄不垫价格数字。最终由 `theme-tokens.test.js` 对比度计算把关。

## 任务（纵切，按依赖顺序）

### Phase 1 — 地基
- [x] **T1** `app.wxss` 令牌重映射到 JFW + 新增 5 个 `--jf-*` + 重写 `theme-tokens.test.js`（值 + 重定阈值）— S ✅ (176 绿；FG-2→#6E646A、price→#D11414、green-2 槽位复用为深粉)
- [x] **T2** tabbar 激活色 `#234B3A`→`#FF4896`（3×SVG + `.ctab-t_on`）+ `custom-tabbar.test.js` — S ✅
- [x] **Checkpoint A**：`npx jest theme-tokens custom-tabbar` 绿 + 全局变粉 ✅

### Phase 2 — 各屏
- [x] **T3** 菜单页：价格红(令牌已粉)/ 针+空碗+饮品插画描边改粉 / cartbar 渐变→深粉 / 暖瓦底→粉 + `menu-reskin.test.js` — M ✅ (优惠圈属新组件需数据，本期不加；菜单为左轨非#标题)
- [x] **T4** 首页：按 screens-shot 重设计 — 粉渐变头 / 分段图标粉 / nav 粉 / 现炸 banner(炸鸡插画) / `#`人气推荐 / 招牌窑鸡·黄金鸡米花促销 + 品牌改鸡福旺(logo 鸡、tagline、json 标题) + `home-reskin`/`nav-color` 测试 — M ✅ (新增炸鸡腿/鸡米花 SVG,已 chrome 验证)
- [x] **T5** 结算页：内联 switch 色 #234B3A→#FF4896 / `#`深蓝板块标题(订单明细·支付方式) / 暖瓦底→粉 — S ✅ (令牌已驱动应付红/确认支付粉；微信绿 #1AAD19 徽章有意保留)
- [x] **T6** 我的页：头/卡/tab 已由令牌驱动为粉；修复收支对比度回归(收入→深粉 4.88、支出/过期警示→价格红 5.41) — M ✅ (金色余额卡因金字 on 金底失守 WCAG 放弃，保留品牌粉卡白字)
- [x] **Checkpoint B**：`npx jest` 全量绿(177) + 四屏已 reskin ✅

### Phase 3 — 清扫 & QA
- [x] **T7** 残留色 grep 干净(仅留微信绿徽章) + app/各页 nav #F3EEE4→#FBEFF3 + app/login 标题→鸡福旺 + invite/share-code/login 全令牌驱动已自动变粉 — S/M ✅
- [x] **Checkpoint C**：全量绿(181) + 残留 grep 干净 ✅ | 真机逐屏对照效果图 ✅(weapp-dev 注入数据截图：首页/菜单体/结算体均与稿一致)

## 守护测试映射（改代码即同步）
- T1 → `theme-tokens.test.js`
- T2 → `custom-tabbar.test.js`
- T3 → `menu-reskin.test.js`
- T4 → `home-reskin.test.js` + `nav-color.test.js`

## 跟进（T3/T5 仅改色，未对齐版式 → 补结构）
- [x] **F1** 菜单页对齐 spec-shot：加品牌海报头(#鸡福旺+现炸出炉+标语药丸) + `#`板块标题 + 黄底红字价格标 + body 高度让位海报 — ✅ (chrome 验证 verify-menu.png；优惠圈需 per-item 数据,未加)
- [x] **F2** 结算页对齐 screens-shot：应付总额 text-primary 粉→价格红(与提交栏一致) — ✅ (verify-oc.png)
- [x] **F3** 主操作按钮去微信原生绿：删除 `<button type="primary">`(原生绿盖不掉作者样式)，只留 weui-btn_primary 类→粉。真机核对 确认支付/扫码点餐 = rgb(255,72,150) ✅ + primary-button-brand.test.js
- [x] **F4** 菜单卡片浮起：扁平分隔列表→淡粉底浮白圆角卡(圆角+粉调柔投影+卡间留白) + 去结算 白字→深蓝字(对比度+对齐稿) — 真机 dev-menu-body2.png ✅

## mock-screens.html 对齐（2026-06-28 第二轮）
- [x] **G1** Baloo 2 数字字体：2.8KB base64 ttf 子集(wght700, 0-9 . , + - ¥ %)内嵌 app.wxss @font-face + `--font-number` 令牌 → home/menu/order-confirm/profile 价格金额。真机度量验证 `.menu-price "¥88.00"`=67.11px == Baloo 字宽(系统回退会不同) ✅ + number-font.test.js
- [x] **G2** 首页：logo 黄底 / 「更多›」蓝色#0066FF+700 / promo 菜名 800 / 扫码框浅粉底+深粉码 ✅ + home-reskin.test.js
- [x] **G3** 结算：抵扣行 label+金额深粉#D81A60(亮粉 -¥ 失守→深粉 4.96:1) / 微信徽章 #1AAD19→#07C160 / 应付总额 36→40rpx ✅ + order-confirm-reskin.test.js
- [x] **G4** 我的：粉渐变页头+黄头像白字 / 下划线式 tab(粉字900+粉条) / 金色余额卡折中(--jf-gold-ink #8A5500 on 金底 5.1:1 过 AA) ✅ + profile-reskin.test.js + theme-tokens 对比度把关
- 真机核对：本轮 mp_screenshot 全程超时，改用 element_getStyles 计算值核对(头像黄/卡金渐变/tab粉/页头粉渐变/昵称白 全部确认) + Baloo 字宽度量。全量 208 测试绿。
- 未做(需决策/资源，用户本轮未点)：庆科黄油体中文显示字(体积大,放弃用系统800模拟) / 彩色食物缩略图(待 OSS 真图) / profile 余额卡上提压头版式

## 真机核对结论（spec-shot/screens-shot 逐屏）
- 首页 / 菜单体 / 结算体：版式·配色·令牌全部与稿一致。
- 仍欠（依赖资源/数据，非本期）：① 缩略图真实菜品摄影(现为粉色占位块)；② 卡片左上 9.9 优惠圈(需 per-item promo 数据)。

## 不在本期
- 真实菜品摄影（OSS 后填，插画/占位过渡）；优惠券/自提/预约；新增屏
