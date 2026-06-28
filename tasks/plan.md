# 实施计划：规格选择弹层重构 + 购物车弹层

视觉/交互真源：`~/Downloads/mp-cart.png`（左屏=规格弹层，右屏=已选商品弹层）。
**配色用已上线的鸡福旺(JFW)粉/黄令牌，不用参考图的墨绿/陶土橙。**
（上一轮 JFW reskin 计划已归档至 `tasks/plan.archived-jfw-reskin.md`。）

## Overview
菜单页两个前端功能（不动后端、不动数据库）：
1. **规格选择弹层重构** — 改成参考图版式：商品信息头 + 规格单选药丸 + 数量步进 + 整宽「加入购物车」。取代现有“每个规格一行各带『加入』按钮”的列表。
2. **购物车弹层** — 点购物车摘要区弹出「已选商品」底部面板，对已加入的菜品改数量/清空；「去结算」仍跳 order-confirm。

## 架构决策（已与用户确认）
- **扁平单维规格**（后端 `ProductSpec` 一个规格=一个价，无温度×糖度 SKU 矩阵）→ 弹层只有一个“规格”药丸组；多属性定制不在本期（需另起后端 SKU epic）。
- **复用客户端购物车 API**（`addToCart`/`updateCartQuantity`/`clearCart`/`getCart`，`utils/storage.js`）原样不动 → `cart-sku.test.js` 等 API 行为测试保持绿；改动只在菜单页交互层（`menu/index.js` + wxml/wxss）。
- **JFW 令牌**：粉 `--weui-BRAND`、黄 `--accent`、深蓝 `--jf-title-blue`、深粉 `--green-2`、`--font-number`。主按钮用 `weui-btn weui-btn_primary` 类（**不带 `type="primary"`**，原生绿盖不掉，见 [[weui-primary-button-green]]）。
- **去结算**：弹层/底栏的去结算仍 `goCart()` 跳 `order-confirm`（带 shop_id/table_no/order_type）。
- **分支**：`feat/spec-cart-sheet`（从 main 切；main 是已上线 reskin）。
- 两个弹层复用同一套“底部上滑 sheet + 遮罩”样式，风格统一。

## 依赖图
```
客户端购物车 API（已存在，不改）
  └── updateCartInfo()  [扩展：同时构建 cartItems[] 供弹层渲染]
        ├── 规格弹层    (T1 逻辑 → T2 版式)
        └── 购物车弹层  (T3 逻辑 → T4 版式)
JFW 令牌（已存在） ── 两个弹层的样式
```
两功能相互独立；先做规格弹层（主加购路径），再做购物车弹层（编辑规格弹层加进来的东西）。

## 任务（纵切，按依赖顺序）

### Phase 1 — 规格选择弹层重构
- **T1 规格弹层交互逻辑（`menu/index.js` + `utils/spec.js`）— S/M**
  - `openSpecPicker`：默认选中首个在售规格(`status===1`)，`specQty=1`。
  - `selectSpec(specId)`：切换选中（售罄不可选）；`specDec/specInc`：`specQty ±1`，下限 1。
  - `confirmAddSpec`：`addToCart(product, selectedSpec, specQty)` → `updateCartInfo` → toast → 关闭。删除按行 `pickSpec`。
  - 抽纯函数 `pickDefaultSpec(specs)`、`clampQty(n)` 到 `utils/spec.js` 单测。
  - **验收**：默认选首个在售+数量1 / 切换规格更新选中态与显示价 / 步进不低于1 / 加入按所选规格×数量入车 / 全售罄时禁用加入。
  - **验证**：`npx jest cart-sku spec-picker` 绿；纯函数单测绿。
  - 文件：`menu/index.js`、`utils/spec.js`(新)、`__tests__/spec-picker.test.js`(新)

- **T2 规格弹层版式 + JFW 样式（`menu/index.wxml` + `index.wxss`）— M**
  - 头部：商品名(900) + 描述。规格组：单选药丸（选中=粉填充/白字，售罄置灰）。价格随选中规格。数量步进（复用 `.menu-step-btn` 粉边圆）。整宽「加入购物车」=`weui-btn weui-btn_primary`。底部上滑 sheet + 遮罩点击关闭。
  - **验收**：结构=头/单选药丸组/价+步进/整宽按钮，与参考图左屏一致；全 JFW 令牌；无 `type="primary"`。
  - **验证**：结构测试绿（绑定 `selectSpec/specInc/specDec/confirmAddSpec`、`weui-btn_primary`、JFW 令牌、无 `type="primary"`）；真机 `element_getStyles`/截图对照参考图。
  - 文件：`menu/index.wxml`、`menu/index.wxss`、`__tests__/spec-picker.test.js`

- **Checkpoint A**：`npx jest` 全绿 + 真机规格弹层符合参考图、加购金额/规格正确。

### Phase 2 — 购物车弹层
- **T3 购物车弹层逻辑（`menu/index.js`）— M**
  - `updateCartInfo` 扩展：构建 `cartItems[]`（`key/name/specName/price/image/quantity`）。
  - `openCartSheet/closeCartSheet`（`cartSheetVisible`）；空车自动关闭。
  - `.menu-cart-left` 的 tap 从 `goCart` 改为 `openCartSheet`；`.menu-cart-go` 仍 `goCart`。
  - `cartDec/cartInc(key)`：`updateCartQuantity` → `updateCartInfo`（减到 0 删行，删空关闭）。`clearCartAll`：`clearCart` → 关闭。
  - **验收**：点摘要区弹面板 / 步进改量底栏总价件数实时同步 / 某行减到0移除、车空关面板 / 清空移除全部并关 / 去结算仍跳 order-confirm。
  - **验证**：`cart-sku`/`cart-isolation` 绿；`cartItems` 构建纯函数单测绿。
  - 文件：`menu/index.js`、`__tests__/cart-sheet.test.js`(新)

- **T4 购物车弹层版式 + JFW 样式（`menu/index.wxml` + `index.wxss`）— M**
  - 标题「已选商品」+「清空」(右上链接)；列表项：缩略图 + 名称/规格 + 价(`--font-number`) + 粉边步进。底部沿用现有浮动粉车条(¥总价 + 黄底去结算)。上滑 sheet + 遮罩。
  - **验收**：结构与参考图右屏一致；JFW 令牌；空态/滚动正常。
  - **验证**：结构测试绿（`cartInc/cartDec/clearCartAll` 绑定、清空链接、JFW）；真机对照。
  - 文件：`menu/index.wxml`、`menu/index.wxss`、`__tests__/cart-sheet.test.js`

- **Checkpoint B**：`npx jest` 全绿（含既有 cart 行为不变）+ 真机两弹层符合参考图，改量/清空/结算闭环。

### Phase 3 — 收尾
- **T5 回归 + 真机闭环 + 合并 — S**
  - 全量 jest 绿；加购→规格弹层→购物车弹层改量→去结算→order-confirm 全链路真机走查（`element_getStyles` 计算值兜底，截图工具偶超时）。
  - 合并 `feat/spec-cart-sheet` → main；按需 DevTools CLI 重新上传小程序（见 [[deploy-topology]]）。

## 风险与缓解
| 风险 | 缓解 |
|---|---|
| 真机 `mp_screenshot` 本会话超时 | 用 `element_getStyles` 计算值 + 必要时人工截图核对（同 reskin 几轮） |
| 弹层与 tabbar/车条 z-index、safe-area 适配 | sheet 设高 z-index + `env(safe-area-inset-bottom)`，复用车条定位经验 |
| 规格弹层与“无规格直接加购”冲突 | 无规格仍走卡片 `+`/步进（`onAdd/onInc/onDec`），弹层只用于有规格商品 |
| 页面 Page 方法难单测 | 把判定/计算抽成 `utils/spec.js` 纯函数单测；交互行为靠结构测试 + 复用已测 cart API + 真机走查 |

## 守护测试映射
- T1 → `spec-picker.test.js`（+ `utils/spec` 纯函数）；`cart-sku` 保持绿
- T2 → `spec-picker.test.js`（结构）
- T3 → `cart-sheet.test.js`；`cart-isolation`/`cart-sku` 保持绿
- T4 → `cart-sheet.test.js`（结构）

## 已定（用户确认 2026-06-28）
- 规格模型 = **扁平，重做交互**（多属性 SKU 不在本期）
- 去结算 = **弹层内跳 order-confirm**
- 分支 = **feat/spec-cart-sheet**

## 不在本期
- 温度×糖度多属性 SKU（需后端属性组/SKU/定价 + 商家后台 + 迁移）
- 真实菜品摄影、优惠圈（既有 deferred）
