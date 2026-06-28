# Todo：规格弹层重构 + 购物车弹层

详见 `tasks/plan.md`。视觉真源：`~/Downloads/mp-cart.png`。
TDD、一任务一提交；改代码同提交内同步它守护的测试。设备核对用 weapp-dev MCP（截图超时则 `element_getStyles` 计算值）。
**git 卫生**：只 stage 该任务文件 + `tasks/`；禁止 `git add -A`；绝不 stage `frontend/config.js`/`.claude/`。
分支：`feat/spec-cart-sheet`（已切）。

## ✅ 已定（2026-06-28）
- [x] Q1 规格模型 = **扁平，重做交互**（多属性 SKU 不在本期）
- [x] Q2 去结算 = **弹层内跳 order-confirm**
- [x] Q3 分支 = **feat/spec-cart-sheet**

## 任务

### Phase 1 — 规格选择弹层重构
- [x] **T1** 规格弹层交互逻辑：`openSpecPicker` 默认选首个在售 + `selectSpec`/`specInc`/`specDec`/`confirmAddSpec`，删按行 `pickSpec`；抽 `utils/spec.js`(`pickDefaultSpec`/`clampQty`/`specPickerState`) + `spec-picker.test.js`(11) 单测 — S/M ✅ (wxml 由 T2 改接；本提交后规格弹层暂以新逻辑待接线)
- [ ] **T2** 规格弹层版式+JFW 样式：商品头 + 单选药丸组 + 价+步进 + 整宽 `weui-btn_primary`（无 type）；底部 sheet+遮罩 + `spec-picker.test.js` 结构 — M
- [ ] **Checkpoint A**：`npx jest` 全绿 + 真机规格弹层符合参考图、加购正确

### Phase 2 — 购物车弹层
- [ ] **T3** 购物车弹层逻辑：`updateCartInfo` 建 `cartItems[]`；`open/closeCartSheet`；`.menu-cart-left` 改 `openCartSheet`；`cartInc/cartDec/clearCartAll`（减0删行、车空关闭）+ `cart-sheet.test.js` — M
- [ ] **T4** 购物车弹层版式+JFW 样式：「已选商品」+「清空」+ 列表项(缩略图/名规/价/粉边步进)；沿用浮动粉车条 + `cart-sheet.test.js` 结构 — M
- [ ] **Checkpoint B**：`npx jest` 全绿（含既有 cart 行为不变）+ 真机两弹层闭环（改量/清空/结算）

### Phase 3 — 收尾
- [ ] **T5** 全量 jest 绿 + 全链路真机走查（加购→规格弹层→购物车弹层改量→去结算→order-confirm）；合并 `feat/spec-cart-sheet`→main，按需重传小程序 — S

## 守护测试映射
- T1 → `spec-picker.test.js`（+ `utils/spec` 纯函数）；`cart-sku` 保持绿
- T2 → `spec-picker.test.js`
- T3 → `cart-sheet.test.js`；`cart-isolation`/`cart-sku` 保持绿
- T4 → `cart-sheet.test.js`

## 不在本期
- 温度×糖度多属性 SKU（需后端 SKU epic）；真实菜品摄影、优惠圈（既有 deferred）
