# Todo：优化「外卖」入口白屏/加载体验

详见 `tasks/plan.md`。范围：**仅 `frontend/`**，不改后端/接口/模型。
根因：外卖比堂食多一段「导航前静默 `resolveDeliveryShop` 请求」+ 冗余 `getShop`。
方案：解析下沉菜单页（点按即跳转）+ DTO 透传跳过 `getShop`（3 请求 → 2）。
TDD：先改测试再改实现，一任务一提交。
**git 卫生**：只 stage 该任务文件；禁止 `git add -A`；绝不 stage `frontend/config.js`/`.claude/`。
分支：建议 `feat/delivery-fast-entry`（开工前切）。

## 任务
- [x] **T1** 菜单页承接外卖冷启动（`onLoad` 无 shop_id 分支 + `loadDeliveryShop` + `loadData(prefetchedShop)` 跳过 `getShop` + `onRetry` 分流）— M ✅
- [ ] **T2** 首页「外卖」即时跳转（`chooseDelivery` 直接 `reLaunch?order_type=delivery`，移除未用导入；无门店提示迁至菜单）— S
- [ ] **Checkpoint A** `cd frontend && npm test` 全绿 + 手测三条（外卖即时/无门店 error/堂食不回归）+ git diff 仅 `frontend/` + 部署停等用户

## 守护测试映射
- T1 → `__tests__/menu-page.test.js` 或新增 `__tests__/delivery-cold-start.test.js`
- T2 → `__tests__/home-launcher.test.js`（改写为即时导航；删「首页无门店提示」用例，意图迁 T1）

## 不在本期
- 后端合并端点、骨架屏、`navigateTo` 转场、预拉取/缓存；任何后端/模型/接口改动。
