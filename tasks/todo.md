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
- [ ] **T3** 菜单页：黄底红字价签 / `#`深蓝标题 / 优惠圈 / 粉加购 / 针+插画描边改粉 + `menu-reskin.test.js` — M
- [ ] **T4** 首页：粉渐变头 / 分段+图标改粉 / nav 色 / banner / 促销 + **品牌文案改鸡福旺(四月春膳→鸡福旺、tagline、logo 春→鸡、json 标题)** + `home-reskin.test.js`(含 `/四月春膳/`→`/鸡福旺/`) + `nav-color.test.js` — M
- [ ] **T5** 结算页：门店卡 / `#`明细 / 应付红 / 福利金开关 / 确认支付 + 内联 switch 色 — S
- [ ] **T6** 我的页：粉会员头 / 金余额 / 三栏 / tab 激活粉 / 订单卡 — M
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
