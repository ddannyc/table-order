# DESIGN.md — 小程序设计规范

视觉规范的**唯一真源**。所有 UI 改动必须遵循本文件。
配色取自 `ui-add-food.png`（**暖橙**：单一品牌橙 + 金黄→桃橙顶部渐变 + 深青墨文字），布局参考同图（左分类栏 + 右菜品列表）。

> 落地方式：颜色以 CSS 变量（design token）定义在 `frontend/app.wxss` 的 `page {}` 中，组件只引用变量，不写死十六进制色值。本文件的 token 名与 `app.wxss` 一致，改色即改这里。导航栏色是唯一例外（小程序 JSON 不支持变量），写在各页 `.json`。

---

## 1. 配色 Color

### 品牌色 Brand（暖橙，取样自 ui-add-food.png）
图中只有**一种**强调橙：选中分类、选规格徽标、底部选中 Tab、加购/CTA 都用它。

| Token | 值 | 用途 |
|------|------|------|
| `--brand-primary` | `#F88818` | 主色（导航栏、选中态、主按钮、选规格徽标、加购、CTA） |
| `--brand-primary-dark` | `#E0760A` | 主色按下/深色态 |
| `--brand-primary-light` | `rgba(248,136,24,0.12)` | 主色浅底（选中分类底、标签底） |
| `--brand-accent` | `#F88818` | 强调（与主色同一橙；保留 token 兼容组件引用） |
| `--brand-accent-dark` | `#E0760A` | 强调按下态 |
| `--brand-accent-light` | `rgba(248,136,24,0.12)` | 强调浅底 |
| `--brand-gradient` | `linear-gradient(135deg, #F6C84A 0%, #F89A70 100%)` | 顶部 banner（金黄→桃橙），仅用于启动页 hero / 菜单门店栏 |

> 迁移说明：曾用 teal `#189CA8` / 微信绿 `#07c160` 均已废弃。`navigationBarBackgroundColor` 全部为 `#F88818`，`navigationBarTextStyle: white`。

### 文本 Text
| Token | 值 | 用途 |
|------|------|------|
| `--color-text-primary` | `#083038` | 主文本/标题（深青墨，取自图片深色） |
| `--color-text-secondary` | `rgba(8,48,56,0.60)` | 次要文本、规格描述 |
| `--color-text-hint` | `rgba(8,48,56,0.35)` | 占位、辅助提示 |
| `--color-text-disabled` | `rgba(8,48,56,0.20)` | 不可用 |
| `--color-text-on-brand` | `#FFFFFF` | 橙底上的文字 |

### 中性 & 功能 Neutral / Functional
| Token | 值 | 用途 |
|------|------|------|
| `--color-bg-page` | `#F5F5F5` | 页面底 |
| `--color-bg-surface` | `#FFFFFF` | 卡片/列表底 |
| `--color-bg-rail` | `#F5F5F5` | 左侧分类栏未选中底 |
| `--color-bg-hover` | `#FAF2E8` | 悬浮/按压暖底 |
| `--color-border` | `rgba(8,48,56,0.08)` | 分割线、边框 |
| `--color-border-light` | `rgba(8,48,56,0.04)` | 细分割线 |
| `--color-price` | `#083038` | 价格用深墨（**忠于图片**；如需更醒目可改 `#F88818`） |
| `--color-success` | `#2BA471` | 成功（绿色，语义更准） |
| `--color-danger` | `#E5484D` | 错误/售罄/删除 |

### 装饰 Decorative（弱使用，仅 banner/插画/空状态）
暖色系：金黄 `#F6C84A`、桃橙 `#F89A70`、浅杏 `#F8D8B0`。顶部 banner 用 `--brand-gradient`。

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
- 阴影：卡片 `0 2rpx 12rpx rgba(8,48,56,0.06)`；底部条 `0 -2rpx 12rpx rgba(8,48,56,0.06)`。

---

## 4. 组件规范 Components

### 按钮
- **主按钮**：底 `--brand-primary`，文字白，圆角 `--radius-lg`，按下 `--brand-primary-dark`。
- **强调/加购**：橙底 `--brand-accent`；「选规格」用橙底白字小徽标（圆角 `--radius-lg`）。
- **次按钮**：白底 + `--brand-primary` 描边 + 主色文字。

### 堂食 / 外卖 切换（toggle）
胶囊分段控件，选中段 `--brand-accent` 橙底白字，未选中透明 + 次要文字。出现在：首页两入口卡片、菜品页顶部。

### 分类栏（左）+ 菜品列表（右）
- 左栏宽约 `180rpx`，底 `--color-bg-rail`；选中项白底 + 左侧 `6rpx` 主色竖条 + 主色文字加粗（亦可用 `--brand-primary-light` 暖底，贴近图片填充观感）。
- 右栏：菜品行 = 左图（圆角 `--radius-md`）+ 名称/规格/价格 + 右下加购控件；价格用 `--color-price` 深墨。

### 价格
统一 `--color-price`（深墨，忠于图片），`¥` 符号小一号。

### 顶部 banner
启动页 hero、菜单门店栏顶部用 `--brand-gradient`（金黄→桃橙），白字。

### 底部购物车条 / TabBar
- 购物车结算条：深色底 + 右侧「去结算」橙色 `--brand-accent` 强调。
- 自定义 TabBar 选中态：图标+文字用 `--brand-primary`。

---

## 5. 应用清单（全量改色范围）
导航栏、TabBar、首页、菜品页、订单确认、邀请、我的、登录、分享码 — 全部遵循本规范。组件不得出现 `#189CA8`、`#07c160` 或写死色值，一律引用 token（导航栏 JSON 色除外，统一 `#F88818`）。
