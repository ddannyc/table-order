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
- [ ] **Checkpoint A**：`npx jest theme-tokens custom-tabbar` 绿 + 全局变粉 + 人工评审

### Phase 2 — 各屏
- [x] **T3** 菜单页：价格红(令牌已粉)/ 针+空碗+饮品插画描边改粉 / cartbar 渐变→深粉 / 暖瓦底→粉 + `menu-reskin.test.js` — M ✅ (优惠圈属新组件需数据，本期不加；菜单为左轨非#标题)
- [x] **T4** 首页：按 screens-shot 重设计 — 粉渐变头 / 分段图标粉 / nav 粉 / 现炸 banner(炸鸡插画) / `#`人气推荐 / 招牌窑鸡·黄金鸡米花促销 + 品牌改鸡福旺(logo 鸡、tagline、json 标题) + `home-reskin`/`nav-color` 测试 — M ✅ (新增炸鸡腿/鸡米花 SVG,已 chrome 验证)
- [x] **T5** 结算页：内联 switch 色 #234B3A→#FF4896 / `#`深蓝板块标题(订单明细·支付方式) / 暖瓦底→粉 — S ✅ (令牌已驱动应付红/确认支付粉；微信绿 #1AAD19 徽章有意保留)
- [x] **T6** 我的页：头/卡/tab 已由令牌驱动为粉；修复收支对比度回归(收入→深粉 4.88、支出/过期警示→价格红 5.41) — M ✅ (金色余额卡因金字 on 金底失守 WCAG 放弃，保留品牌粉卡白字)
- [ ] **Checkpoint B**：`npx jest` 全量绿 + 四屏贴合效果图 + 人工评审

### Phase 3 — 清扫 & QA
- [ ] **T7** 残留色 grep 清扫 + app.json nav + invite/share-code/login 抽查 + 全设备 QA — S/M
- [ ] **Checkpoint C**：全量绿 + 残留 grep 干净 + 真机贴合 → 可提交/开 PR

## 守护测试映射（改代码即同步）
- T1 → `theme-tokens.test.js`
- T2 → `custom-tabbar.test.js`
- T3 → `menu-reskin.test.js`
- T4 → `home-reskin.test.js` + `nav-color.test.js`

## 不在本期
- 真实菜品摄影（OSS 后填，插画/占位过渡）；优惠券/自提/预约；新增屏
