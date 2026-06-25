# DESIGN.md — 小程序设计规范

视觉规范的**唯一真源**。所有 UI 改动必须遵循本文件。
配色取自 `ui-color.png`（teal 主色 + orange 强调），布局参考 `ui-add-food.png`（左分类栏 + 右菜品列表）但按本规范重新配色。

> 落地方式：颜色以 CSS 变量（design token）定义在 `frontend/app.wxss` 的 `page {}` 中，组件只引用变量，不写死十六进制色值。本文件的 token 名与 `app.wxss` 一致，改色即改这里。

---

## 1. 配色 Color

### 品牌色 Brand
| Token | 值 | 用途 |
|------|------|------|
| `--brand-primary` | `#189CA8` | 主色（导航栏、主按钮、选中态、价格强调外的主要强调） |
| `--brand-primary-dark` | `#0C7A85` | 主色按下/深色态 |
| `--brand-primary-light` | `rgba(24,156,168,0.12)` | 主色浅底（选中分类底、标签底） |
| `--brand-accent` | `#FC8400` | 强调橙（价格、选规格徽标、加购、活动标签、次级 CTA） |
| `--brand-accent-dark` | `#E07600` | 橙按下态 |
| `--brand-accent-light` | `rgba(252,132,0,0.12)` | 橙浅底 |

> 迁移说明：原全局 `--weui-primary: #07c160`（微信绿）→ 改为 `#189CA8`。`navigationBarBackgroundColor` 同步改为 `#189CA8`。

### 文本 Text
| Token | 值 | 用途 |
|------|------|------|
| `--color-text-primary` | `#00303C` | 主文本/标题（深青墨，取自 palette 深色） |
| `--color-text-secondary` | `rgba(0,48,60,0.60)` | 次要文本、规格描述 |
| `--color-text-hint` | `rgba(0,48,60,0.35)` | 占位、辅助提示 |
| `--color-text-on-brand` | `#FFFFFF` | 主色/橙色底上的文字 |

### 中性 & 功能 Neutral / Functional
| Token | 值 | 用途 |
|------|------|------|
| `--color-bg-page` | `#F5F7F7` | 页面底（带极浅青调的灰白） |
| `--color-bg-surface` | `#FFFFFF` | 卡片/列表底 |
| `--color-bg-rail` | `#F2F4F5` | 左侧分类栏未选中底 |
| `--color-border` | `rgba(0,48,60,0.08)` | 分割线、边框 |
| `--color-price` | `#FC8400` | 价格统一用橙（与 `ui-add-food.png` 一致） |
| `--color-success` | `#189CA8` | 成功（沿用主色，不再用微信绿） |
| `--color-danger` | `#E5484D` | 错误/售罄/删除 |

### 装饰 Decorative（弱使用，仅插画/空状态/渐变）
`#84CCCC`（浅青）、`#A8D8FC`（浅蓝）、`#C0E4C0`（浅绿）、`#F0B46C`（暖棕）。
顶部 banner 可用主色渐变：`linear-gradient(135deg, #189CA8, #18909C)`。

---

## 2. 字号 Typography（沿用现有 rpx 体系）
| Token | 值 | 用途 |
|------|------|------|
| `--font-caption` | `24rpx` | 规格、辅助 |
| `--font-body` | `28rpx` | 正文 |
| `--font-subhead` | `30rpx` | 列表标题 |
| `--font-title` | `36rpx` | 菜品名、区块标题 |
| `--font-h2` | `40rpx` | 页面副标题 |
| `--font-h1` | `48rpx` | 页面主标题 |
| `--font-price` | `34rpx` | 价格 |

字重：标题 `600`，正文 `400`，价格 `600`。字体沿用系统栈（`app.wxss` 已定义）。

---

## 3. 间距 / 圆角 / 阴影
- 间距：`--space-xs 8` / `--space-sm 16` / `--space-md 24` / `--space-lg 32` / `--space-xl 48`（rpx）。
- 圆角：`--radius-sm 8` / `--radius-md 16` / `--radius-lg 24`；按钮与徽标用 `--radius-lg`（参考图圆角偏大）。
- 阴影：卡片 `0 2rpx 12rpx rgba(0,48,60,0.06)`；底部条 `0 -2rpx 12rpx rgba(0,48,60,0.06)`。

---

## 4. 组件规范 Components

### 按钮
- **主按钮**：底 `--brand-primary`，文字白，圆角 `--radius-lg`，按下 `--brand-primary-dark`。
- **强调/加购**：橙底 `--brand-accent` 或橙描边；「选规格」用橙底白字小徽标（圆角 `--radius-lg`）。
- **次按钮**：白底 + `--brand-primary` 描边 + 主色文字。

### 堂食 / 外卖 切换（toggle）
胶囊分段控件，选中段 `--brand-accent` 橙底白字，未选中透明 + 次要文字（参考图右上角样式）。出现在：首页两入口卡片、菜品页顶部。

### 分类栏（左）+ 菜品列表（右）
- 左栏宽约 `180rpx`，底 `--color-bg-rail`；选中项白底 + 左侧 `6rpx` 主色竖条 + 主色文字加粗。
- 右栏：菜品行 = 左图（圆角 `--radius-md`）+ 名称/规格/价格 + 右下加购控件；价格用 `--color-price` 橙色。

### 价格
统一橙色 `--color-price`，`¥` 符号小一号。

### 底部购物车条 / TabBar
- 购物车结算条：主色 `--brand-primary` 底，右侧「去结算」可用橙色强调。
- 自定义 TabBar 选中态：图标+文字用 `--brand-primary`（替换原微信绿）。

---

## 5. 应用清单（全量改色范围）
导航栏、TabBar、首页、菜品页、订单确认、邀请、我的、登录、分享码 — 全部从微信绿迁移到本规范。组件不得出现 `#07c160` 或写死色值，一律引用 token。
