# Todo：修复菜单「选规格」按钮三位数价格换行（方案 V3）

详见 `tasks/plan.md`。范围：**仅 `frontend/`**（utils 价格函数 + menu WXML/WXSS + 测试），不改接口/业务逻辑/config.js。
方案 V3（效果图 375px 真机比例验证）：瘦黄标 + 整数去 .00 + `选规格` 紧凑原子药丸 → `¥38/¥100/¥288/¥1288 起` 全档同行不断字；`flex-wrap` 仅兜底。
TDD：先写测试再改；一任务一提交。
**git 卫生**：只 stage 该任务文件；禁止 `git add -A`；绝不 stage `frontend/config.js`/`.claude/`。
分支：建议 `feat/menu-price-btn-fit`（开工前切；注意当前在 feat/menu-skeleton，需先理清分支）。

## 任务
- [x] **T1** 价格格式化纯函数 `utils/price.js formatPrice`（整数去 .00 / 非整保两位）+ 接入 menu loadData（`priceText`/`specMinText`）+ 单测 — S ✅
- [ ] **T2** 价格标瘦身 + `选规格` 紧凑原子（`menu-spec-btn`：nowrap+flex-shrink:0）+ `.menu-action{flex:none;margin-left:auto}` + `.menu-card-bottom{flex-wrap:wrap}` + WXML 加类 + 守护测试 — S
- [ ] **Checkpoint A** `npm test` 全绿 + 真机多档价格手测（两/三/四位数 + ¥38.50 + 带徽章均同行不断字）+ git diff 仅 frontend + 部署停等用户

## 守护测试映射
- T1 → `__tests__/price-format.test.js`（整数/非整/0/字符串容错）；`menu-page.test.js` 守护 loadData 不回归
- T2 → `__tests__/menu-price-btn.test.js`（WXML 含 `menu-spec-btn`；WXSS `.menu-spec-btn` 含 nowrap+flex-shrink:0；`.menu-card-bottom` 含 flex-wrap:wrap）

## 不在本期
- 购物车/下单页/规格弹层价格格式统一、缩小缩略图（V4 备选）、改价格标色彩、组件测试栈。

## 已定方向
- 方案 V3（瘦黄标 + 整数去 .00 + 紧凑按钮），全价格档同行不断字；wrap 仅极端兜底。
