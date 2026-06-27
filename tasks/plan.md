# Plan：鸡福旺 配色落地（Pine-Ink → JFW Brand Reskin）

> Status: **READY — 4 个 Open Q 已定（见 Resolved Decisions），未改任何代码，待发话开始 T1。**
> 视觉真源：鸡福旺官方规范（聊天中给定）+ 已认可效果图
> `~/Downloads/spec-shot.png`（菜单）、`~/Downloads/screens-shot.png`（首页/结算/我的）。
> 效果图 HTML 在会话 scratchpad：`mock-spec.html`、`mock-screens.html`。

## 基线（沿用上一轮约定）
- **TDD、一任务一提交**：每个任务在同一提交内改代码 + 同步它守护的测试，绝不让套件跨 checkpoint 变红。
- **git 卫生**：只 `git add` 该任务文件 + `tasks/`；**禁止 `git add -A`**；绝不 stage `frontend/config.js` / `.claude/` / `specs/`。
- **设备核对现已可行**：本轮可用 weapp-dev MCP（`mp_ensureConnection`→`mp_navigate`→`mp_screenshot`）逐屏截真机/开发者工具，与效果图对照——不再只靠"人工自查"。

## Overview

把点餐小程序从 **松墨 Pine-Ink**（奶油 + 墨绿 + 陶土）换成 **鸡福旺 (JFW)** 品牌色
（糖果粉 + 亮黄 + 钴蓝点缀 + 价格红）。代码是**令牌驱动**的：各页大量使用 `var(--weui-BRAND)`、
`var(--accent)`、`var(--price-ink)`、`var(--weui-BG-*)`、`var(--weui-FG-*)`，所以**改 `app.wxss`
一处令牌即可把品牌色级联到全部屏**。规范新增的屏内"招牌做法"（黄底红字价格标、`#`深蓝板块标题、
粉底红字优惠圈）**无对应令牌**，按页落地。

一套**测试硬断言了 Pine-Ink 色值与 WCAG 阈值**；把这些测试改到新调色板是工作的一部分，不是附带项。

## 品牌令牌（目标值，保留 weui 名以继续给 weui 组件上色）

| 令牌 | 旧(Pine-Ink) | 新(JFW) | 说明 |
|---|---|---|---|
| `--weui-BRAND` | `#234B3A` | `#FF4896` | 主粉：按钮/激活态/cartbar/徽章；`!important` 必须保留 |
| `--accent` | `#C8643C` | `#FFD300` | 亮黄：价格标底/装饰 |
| `--price-ink` | `#B0491F` | `#D11414` | 价格红（规范 #E62121 微调：on 卡 ≈5.4:1 过 AA；规范 #E62121 on 白仅 4.39） |
| `--weui-BG-0` | `#F3EEE4` | `#FBEFF3` | 页面·淡粉（已定） |
| `--weui-BG-1` | `#FBF8F2` | `#FFFDF8` | 卡片底（规范 `--jf-card-bg`） |
| `--weui-BG-2` | `#EFE9DD` | `#FCE7EF` | 左类目轨底 |
| `--weui-FG-0` | `#2A2723` | `#222222` | 正文（规范 `--jf-text-black`） |
| `--weui-FG-2` | `#6E665A` | `#8A7F86` | 次要文字（须过 AA on card；Task 1 选值时验证） |
| `--weui-FG-3` | `#E3DCCE` | `#E8E8E8` | 分割线（规范 `--jf-line-gray`） |
| `--green-2` → `--pink-deep` | `#2F6B4F` | `#D81A60` | **承载小号浅色文字的粉带**（白字 on 它 ≈4.96:1 过 AA）；渐变深端 |
| **新增** `--jf-blue` | — | `#0066FF` | 蓝点缀（饮品/链接/边框线） |
| **新增** `--jf-title-blue` | — | `#0A2463` | `#` 板块标题 |
| **新增** `--jf-tag-pink` | — | `#FFE0ED` | 优惠圈底 |
| **新增** `--jf-tag-red` | — | `#C2185B` | 优惠圈字（规范 #D81A60 微调：on tag-pink ≈4.8:1 过 AA） |
| **新增** `--jf-orange` | — | `#FFB829` | 饮品/果茶配图 |

> **粉的双值规则（WCAG 关键）**：`--weui-BRAND #FF4896`（亮粉）只用于**填充/图标/大字**；
> 大字 on 亮粉 ≈3.2:1 满足 **AA-large 3:1**（≥18px 或加粗的正确阈值，非放宽）。**小号浅色文字**
> 一律落在 `--pink-deep #D81A60` 上（白字 ≈4.96:1）。**价签**取消"黄底红字"（仅 3.2:1），改为
> **红字 on 卡**（`--price-ink #D11414` ≈5.4:1）；亮黄 `--accent` 留给优惠圈环/装饰，不垫在价格数字下。
> 以上 nudged 值为 T1 工作值，最终由 `theme-tokens.test.js` 的对比度计算把关。

## 架构决策
1. **重映射既有 weui 令牌名，而非全面引入 `--jf-*`**：改动最小，`var(--weui-BRAND)` 已贯穿各页与 weui `type="primary"` 按钮，翻一处声明即全局变粉。仅给规范新增色补 `--jf-*`。
2. **测试即契约**：每处 Pine-Ink 断言（`theme-tokens`/`custom-tabbar`/`home-reskin`/`menu-reskin`/`nav-color`）在**改它所守护代码的同一任务里**同步更新到 JFW。
3. **WCAG：保留 4.5:1 阈值，微调色值达标**（已定 = Open Q1）。不放宽测试：所有**小号文字**组合过 4.5:1——靠"粉的双值规则"（亮粉只填充/大字，小字落 `--pink-deep`）、价签改红字 on 卡、`--jf-tag-red` 调深。大字 on 亮粉走 AA-large 3:1（≥18px/加粗的**正确**阈值，非放宽）。`theme-tokens.test.js` 的对比度计算最终把关。
4. **按屏纵切**：地基（令牌 + tab）之后，每屏一个自洽任务：改样式 + 改它的守护测试 + 真机截图，完成后 app 仍可用。

## 依赖图
```
app.wxss 令牌块 ──┬──────────────────────────────────────────────┐
(Task 1)+令牌测试  │ （经 var 把品牌色级联到所有页）                │
                  │  菜单      首页       结算         我的         │
组件 tabbar       │ (Task 3)  (Task 4)*  (Task 5)    (Task 6)     │
(Task 2)+tab测试   │   │        │          │           │           │
        └── 共享 ──┴───┴────────┴──────────┴───────────┴───────────┘
                                 │
                      残留色清扫 + 设备 QA (Task 7)
*Task 4 兼管 nav 色（home/index.json）+ nav-color.test.js
```
顺序自底向上：令牌 → tab → 各屏 → QA。

## Task List

### Phase 1 — 地基（共享，阻塞全部）

#### Task 1: `app.wxss` 令牌重映射 + 重写 `theme-tokens.test.js` — S
**描述**：把 Pine-Ink 令牌值换成 JFW（见上表），新增 5 个 `--jf-*`，保留 `--weui-BRAND` 的 `!important`。重写令牌值断言并按决策 #3 重定 WCAG 阈值。
**验收**：
- [ ] `app.wxss` 各令牌为 JFW 值；`--weui-BRAND: #FF4896 !important` 在位。
- [ ] `theme-tokens.test.js` 断言新值 + 重定阈值（浅卡上深字 ≥4.5:1；正文 ≥7:1；白字-on-粉 & 红字-on-黄 ≥3:1）。
- [ ] `app.wxss` 内无 Pine-Ink 残留色（`#234B3A/#C8643C/#B0491F/#2F6B4F`）。
**验证**：`cd frontend && npx jest theme-tokens` 过；开发者工具看菜单页主按钮/cartbar/徽章变粉。
**依赖**：无。**文件**：`frontend/app.wxss`、`frontend/__tests__/theme-tokens.test.js`。

#### Task 2: tabbar 改色 + `custom-tabbar.test.js` — S
**描述**：`components/tabbar/index.wxss` 里 3 个 SVG data-uri 的激活 stroke 与 `.ctab-t_on` 颜色 `#234B3A`→`#FF4896`；未激活 `#8A8275` 不变。更新守护测试期望。
**验收**：
- [ ] 激活图标+文字 `#FF4896`；未激活不变；无 `#07c160`。
- [ ] `custom-tabbar.test.js` 期望 `#FF4896`。
**验证**：`npx jest custom-tabbar` 过；各页底 tab 激活项为粉。
**依赖**：Task 1。**文件**：`frontend/components/tabbar/index.wxss`、`frontend/__tests__/custom-tabbar.test.js`。

### Checkpoint A — 地基
- [ ] `npx jest theme-tokens custom-tabbar` 绿。
- [ ] 开发者工具内全局品牌色为粉（按钮/cartbar/tab/徽章）。
- [ ] **人工评审后再进各屏。**

### Phase 2 — 各屏纵切

#### Task 3: 菜单页 reskin + `menu-reskin.test.js` — M
**描述**：把规范招牌做法落到 `menu/index.wxss`(+`menu-art.wxss`)：黄底红字价格标、`#`深蓝板块标题、粉底红字优惠圈、粉色加购/选规格/stepper、左轨激活粉、cartbar 粉+黄；把定位针 SVG（`menu/index.wxss:36`）与 `menu-art.wxss` 饮品插画描边由绿改粉。更新 `menu-reskin.test.js`（现断言 `%23234B3A`）。
**验收**：
- [ ] 价签黄底(`--accent`)红字(`--price-ink`)；板块标题 `--jf-title-blue` + 粉 `#`。
- [ ] 针/插画描边为 `#FF4896`（无 `#234B3A`）。
- [ ] 版式贴合 `spec-shot.png`。
**验证**：`npx jest menu-reskin` 过；weapp-dev 截菜单页对照 `spec-shot.png`。
**依赖**：Task 1、2。**文件**：`menu/index.wxss`、`menu/menu-art.wxss`、`__tests__/menu-reskin.test.js`。

#### Task 4: 首页 reskin + nav 色 + `home-reskin.test.js` + `nav-color.test.js` — M
**描述**：`home/index.wxss`（粉渐变品牌头、分段激活粉、扫码卡、banner、`#`人气推荐、促销价签）；`home-art.wxss` 分段图标描边绿→粉；`home/index.json` `navigationBarBackgroundColor` `#234B3A`→`#FF4896`。**品牌文案改名**（决策 #2）：`home/index.wxml` 的 `四月春膳`→`鸡福旺`、`一汤一蔬 · 阅沐春风`→`五指毛桃鸡 · 现炸小食`、`banner-h`→`现炸出炉`、logo `春`→`鸡`；`home/index.json` 标题→鸡福旺。更新 `home-reskin.test.js`（`%23234B3A`→粉 + `/四月春膳/`→`/鸡福旺/`）与 `nav-color.test.js`。
**验收**：
- [ ] 头部粉渐变、分段激活粉白、分段图标粉；品牌文案为鸡福旺。
- [ ] 首页 nav bar `#FF4896`；home 美术/json 无 `#07c160`/`#234B3A`；无残留 `四月春膳`。
- [ ] 贴合 `screens-shot.png`（首页）。
**验证**：`npx jest home-reskin nav-color` 过；weapp-dev 截首页对照。
**依赖**：Task 1、2。**文件**：`home/index.wxml`、`home/index.wxss`、`home/home-art.wxss`、`home/index.json`、`__tests__/home-reskin.test.js`、`__tests__/nav-color.test.js`。

#### Task 5: 结算页 reskin（含内联 switch 色）— S
**描述**：`order-confirm/index.wxss`（门店卡、`#`订单明细、金额拆分应付红、福利金卡、确认支付主按钮粉、提交栏）；`order-confirm/index.wxml` 内联 `<switch color="#234B3A">`→`#FF4896`；若此处有残留绿 `#2C5A45` 一并换。
**验收**：
- [ ] switch + 确认支付 + 抵扣强调 用粉/红令牌；无 `#234B3A`。
- [ ] 贴合 `screens-shot.png`（结算页）。
**验证**：`npx jest` 全量绿（结算无专属套件）；weapp-dev 截结算页对照。
**依赖**：Task 1、2。**文件**：`order-confirm/index.wxss`、`order-confirm/index.wxml`。

#### Task 6: 我的页 reskin — M
**描述**：`profile/index.wxss`（粉渐变会员头、金色福利金余额、三栏统计、tab 激活粉下划线、订单卡、收支金额色）。审 `gold`/`muted` 余额类与 `plus/minus` 收支色在新卡底是否可读。
**验收**：
- [ ] 头部粉、余额金、激活 tab 粉下划线；收支色在卡上清晰。
- [ ] 贴合 `screens-shot.png`（我的）。
**验证**：`npx jest` 全量绿；weapp-dev 截我的页对照。
**依赖**：Task 1、2。**文件**：`profile/index.wxss`。

### Checkpoint B — 核心屏
- [ ] `cd frontend && npx jest` 全量绿。
- [ ] 四屏在开发者工具内逐屏贴合效果图。
- [ ] **人工评审。**

### Phase 3 — 清扫 & QA

#### Task 7: 残留色清扫 + app nav + 全设备 QA — S/M
**描述**：全 `frontend`（排除 `node_modules`/`miniprogram_npm`）grep 残留 Pine-Ink（`#234B3A #2F6B4F #C8643C #B0491F`）与零散绿（`#2C5A45 #1AAD19`），修在范围内文件。定 `app.json` nav bar bg（`#F3EEE4`→粉系，决策 #3）。通用标题 `app.json`/`home/index.json` `navigationBarTitleText 餐饮点餐`、`login/index.wxml` `app-name 餐饮点餐` 一并改鸡福旺（决策 #2）。抽查继承令牌但未重设计的页（invite/share-code/login）观感无破。
**验收**：
- [ ] app 自有 `*.wxss/*.wxml/*.json` 内无 Pine-Ink/零散绿残留。
- [ ] invite/share-code/login 在新调色板下观感连贯。
**验证**：`npx jest` 全量绿；`grep -rE "#234B3A|#2F6B4F|#C8643C|#B0491F|#2C5A45|#1AAD19" pages components app.* | grep -v data:image` 无输出；weapp-dev 截各 tab 页。
**依赖**：Task 3–6。**文件**：范围内残留；`frontend/app.json`（可能）。

### Checkpoint C — 完成
- [ ] 全量 jest 绿；残留 grep 干净。
- [ ] 各屏真机贴合效果图。
- [ ] 可提交 / 开 PR（分支见 Open Q4）。

## Risks and Mitigations
| 风险 | 影响 | 缓解 |
|---|---|---|
| 亮色调失守现有 4.5:1 阈值 | 高 | 决策 #3 微调色值（粉双值/价签红字 on 卡/tag-red 调深）；T1 对比度计算把关 |
| 5 个测试硬编 Pine-Ink 色值 | 高 | 每个守护测试在改其目标的同一任务里同步更新 |
| weui `.wx-root` 把绿 `#07c160` 盖回 `page` | 中 | 保留 `--weui-BRAND` 的 `!important`（theme-tokens 测试守护） |
| 装饰 SVG 美术内嵌绿描边 | 中 | Task 3/4 明列改色；menu/home reskin 测试抓残留 |
| 规范新增色（黄/蓝）无令牌槽 | 中 | Task 1 增 `--jf-*`；各页按屏引用 |
| `--weui-FG-2` 次要文字须在新卡过 AA | 中 | Task 1 选值使 `contrast(FG2, card) ≥ 4.5` 并断言 |

## Resolved Decisions（人工已定 2026-06-28）
1. **WCAG = 微调色值达 4.5:1**：保留严格阈值，不放宽测试；按"粉双值规则"+价签红字 on 卡+`--jf-tag-red` 调深达标（见令牌表下方说明与决策 #3）。
2. **品牌文案 = 一并改成鸡福旺**：随 reskin 把 `home/index.wxml` 的 `brandname 四月春膳`、`brandsub 一汤一蔬·阅沐春风`、`banner-h 四月春膳`、logo 字 `春` 改为鸡福旺 / `五指毛桃鸡 · 现炸小食` / `现炸出炉` / `鸡`；`home/index.json` 与 `app.json` 的 `navigationBarTitleText 餐饮点餐`、`login` 的 `app-name 餐饮点餐` 一并改鸡福旺。**`home-reskin.test.js:18` 的 `/四月春膳/` 断言改为 `/鸡福旺/`**（属 T4）。
3. **页面背景 = 淡粉 `#FBEFF3`**：`--weui-BG-0 = #FBEFF3`；app/nav bar bg 走粉系（见 T7 定 app.json nav）。
4. **分支 = 当前 `feat/shansong-delivery`**：不另开分支，按一任务一提交累加。
