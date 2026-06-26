# Plan：质感升级 首页+菜单（松墨 Pine-Ink）

来源 spec：`docs/ideas/texture-uplift-pine-ink.md`（idea-refine 产出，已锁定全部决策）。
风格沿用上一轮 M1–M5：纵切、一任务一提交、TDD（RED→GREEN→回归 jest）。
**基线**：绝不 git add `frontend/config.js` / `.claude/` / `specs/`；每次提交只 stage 该任务文件，禁止 `git add -A`。

---

## 关于"测试"的诚实说明
这是换肤/视觉任务。可自动化的验收 = ①**WCAG 对比度**（用已知 hex 在 jest 里算，硬门槛）②**结构断言**（读 wxml/wxss 字符串断言类名/令牌/SVG 存在，沿用 `menu-page.test.js` 既有写法）③**无行为回归**（既有 jest 全绿）。
**不可自动化** = 实际"质感"观感、真机 SVG data-URI 渲染、亮金描边在真机的颜色 —— 这些放到 C1 checkpoint 人工核对（本环境无模拟器）。计划不假装能测观感。

## 渲染技术决策（地基）
线描插画一律用 **wxss `background-image: url("data:image/svg+xml,<urlencoded>")`** 实现：
- 零二进制资源（承接 M3 决定）、wxss 支持稳、可被结构测试断言。
- 不用 `<image src="data:...svg">`（真机 SVG 支持历史上不稳）。描边色写死金 `#C98A2B`（data-URI 内联无法引 CSS 变量，需在注释标注与令牌同源）。

## 令牌系统（最终值，已过 WCAG）
```
--weui-BRAND  #2C4A3B   主色：品牌band/主按钮/菜单激活竖条/cartbar/徽章底
--weui-BG-0   #F3EEE4   页面·燕麦奶油
--weui-BG-1   #FBF8F2   卡面·近白暖
--weui-BG-2   #EFE9DD   门店栏/左类目轨底
--weui-FG-0   #2A2723   正文（12:1）
--weui-FG-2   #6E665A   次要文字（4.89:1，比参考 #8A8275 调暗以过 AA）
--weui-FG-3   #E3DCCE   暖发丝线
--brand-accent #C98A2B  金：插画描边/分隔/印章（仅填充描边，不做文字）
--price-ink   #A66E1F   价格大字（3.73:1，仅 ≥34rpx 粗体）
```
徽章/数量 chip = 深绿底 + 奶油字（8.44:1）。金绝不做文字。

---

## 依赖图
```
T1 令牌地基 ──┬─→ T2 首页band+入口卡 ──→ T3 首页hero插画
              └─→ T4 菜单重皮 ──┬─→ T5 菜单分类金线glyph
                                 └─→ T6 未绑桌空状态插画
                                            └─→ C1 自审+全量绿+真机核对
```
执行顺序：T1 → T2 → T3 → T4 → T5 → T6 → C1。

---

## T1 — 令牌地基（app.wxss 覆盖 weui 配色 + 新增 accent/price）— S
**改**：`frontend/app.wxss`（仅 `page{}` 令牌 + `.page` 背景；不动既有 spacing/font 令牌与组件类）。
**验收**
- `page{}` 定义上表全部令牌；`.page { background: var(--weui-BG-0) }`。
- 新增 `theme-tokens.test.js`：读 app.wxss 断言关键令牌+hex 存在；并**在 JS 内实现 WCAG 对比度函数**，断言：
  - `--price-ink` on `--weui-BG-0` ≥ 3.0；on `--weui-BG-1` ≥ 3.0
  - `--weui-FG-2` on `--weui-BG-0` ≥ 4.5
  - `--weui-FG-0` on `--weui-BG-0` ≥ 7.0
  - 奶油字(`--weui-BG-0`) on `--weui-BRAND` ≥ 4.5
**验证**：`cd frontend && node node_modules/jest/bin/jest.js` 全绿（无行为回归）。

## T2 — 首页品牌 band + 入口卡重皮 — M（依赖 T1）
**改**：`frontend/pages/home/index.wxml` `.wxss`（结构微调 + 令牌上色；不动 `index.js` 行为）。
- 顶部加深绿品牌 band（字标，仅店名/品牌名，**不含余额/积分**——已决策）。
- 入口卡：`--weui-BG-1` 卡面 + 金发丝线分隔 + 深绿标题；emoji 暂留（hero 在 T3）。
**验收**
- wxml 有 `.home-brandband`；wxss band 背景 `var(--weui-BRAND)`、文字奶油色。
- 入口卡用令牌（无散落硬编码 hex）。
- 既有 `home-launcher.test.js` 三个行为用例**全绿**（堂食扫码、外卖解析、无门店兜底）。
- 新增结构断言：wxml 含 `home-brandband`。
**验证**：jest 全绿。

## T3 — 首页 hero 线描插画（蒸笼一桌菜）— M（依赖 T2）
**改**：`frontend/pages/home/index.wxml` `.wxss`。
- band 下方加 `.home-hero` 横幅，wxss 用 `background-image` data-URI 单线 SVG（蒸笼+筷+碗，金描边），置于入口卡之上。
- 静态（无动画）；如加入场动画须 `@media (prefers-reduced-motion: reduce)` 关闭。
**验收**
- wxml 有 `.home-hero`；wxss `.home-hero` 含 `data:image/svg+xml`。
- `home-launcher.test.js` 仍全绿；新增结构断言 hero data-URI 存在。
**验证**：jest 全绿。

## T4 — 菜单重皮（shopbar/rail/cards/price/cartbar）— M（依赖 T1）
**改**：`frontend/pages/menu/index.wxss`（主要）；wxml 仅在必要处加类（不动结构与 `index.js`）。
- 门店栏→`--weui-BG-2`；卡面→`--weui-BG-1` 抬升；发丝线→`--weui-FG-3`；价格大字→`--price-ink`；`起`后缀→`--weui-FG-2`；cartbar→`--weui-BRAND`+奶油字；数量 chip→深绿底奶油字。激活竖条已用 `--weui-BRAND`（自动变真品牌绿）。
**验收**
- wxss 价格类用 `var(--price-ink)`；cartbar 背景 `var(--weui-BRAND)`；卡面 `var(--weui-BG-1)`。
- `menu-page.test.js` + `cart-isolation.test.js` 全绿（行为/隔离无回归）。
- 新增结构断言：price 令牌 + cartbar 令牌存在。
**验证**：jest 全绿。

## T5 — 菜单分类金线 glyph（升级 M3 占位）— M（依赖 T1、T4）
**改**：`frontend/pages/menu/index.wxss`（按 glyph 加 `.menu-thumb-ph_{cup|bubble|cheese|sparkle}` 的金线 data-URI SVG）；`menu-image.js` 保持（仍供 glyph+label）。
- 占位块从"中性文字"升级为"金色单线 glyph + label 兜底文字"（label 保留供无障碍/识别）。
**验收**
- wxss 定义 4 个 `.menu-thumb-ph_*` 类，各含 `data:image/svg+xml` 金描边。
- `menu-image.test.js` 全绿（解析逻辑不变）；占位仍含 label 文本。
- 新增结构断言：4 个 glyph 类的 SVG 存在。
**验证**：jest 全绿。

## T6 — 未绑桌空状态线描插画 — S（依赖 T4）
**改**：`frontend/pages/menu/index.wxml` `.wxss`（未绑桌 `weui-msg` 区加 `.menu-empty-illu` 金线 SVG）。
- 文案保持主动语态（"扫描餐桌二维码开始点餐"）；插画为单线空碗/纸飞机。
**验收**
- wxml 未绑桌块含 `.menu-empty-illu`；wxss 该类含 data-URI SVG。
- 既有菜单测试全绿；新增结构断言。
**验证**：jest 全绿。

---

## Checkpoint C1 — 设计自审 + 全量绿 + 真机核对
- **/frontend-design 自审**：signature 单一（hero 是唯一亮点，其余克制）；金只做填充描边；激活态不只靠颜色；文案主动语态；prefers-reduced-motion 已尊重。
- **全量 jest 绿**（记录数字）。
- **人工待办（本环境无法做）**：真机核对 ① SVG data-URI 在真机渲染 ② 金/铜对比眼检 ③ 两屏整体观感是否达参考图质感级别。达标后再决定是否铺其余 4 屏。

## 不在本期
- 真实食物摄影；其余 4 屏（下单/我的订单/地址/选店）重皮；自定义衬线/显示字体；类目 scroll-spy；余额/会员积分模块。
