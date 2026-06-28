# Todo：商家后台 · 订单运营台 (Order Action Board)

详见 `tasks/plan.md`。Spec：`specs/merchant-admin.md` §5 缺口⑥ / §9。Idea：`docs/ideas/order-action-board.md`。
范围：后端 `backend/`（PreparedAt + 列表扩展 + 出餐/改状态/重新派单 3 接口 + httptest）与前端 `admin/`（订单运营台）。
**不在本期**：退款/refund、闪送自动重试自愈、派单失败推送告警、堂食/外卖拆视图、看板趋势图。
TDD：后端先写 httptest 再实现；**契约优先**（先把列表 + 3 接口跑绿再做 UI）；一任务一提交。
**git 卫生**：只 stage 该任务文件；禁止 `git add -A`；绝不 stage `frontend/config.js`/`.claude/`。

## Phase 1 — 后端基础（契约）
- [x] **T1** `Order.PreparedAt *time.Time`（出餐时间，非破坏；已确认无处读 `Status==3` 当出餐）+ 迁移 — XS ✅
- [x] **T2** 扩展 `GetMerchantOrders`：LEFT JOIN `order_deliveries` + 返回 `order_type/status/paid_at/prepared_at/delivery{...}` + `status/type` 筛选 + 分页（替换 50 上限）+ 含未支付 + `revenue/rewarded` 按筛选聚合 + httptest — M ✅
- [x] **Checkpoint 1** `go test ./...` 全绿；列表 schema 冻结（前端据此对接）

## Phase 2 — 后端动作接口（可并行）
- [x] **T3** `POST /merchant/orders/:id/prepare`（出餐，置 PreparedAt，校验归属，幂等）+ httptest — S
- [ ] **T4** `PUT /merchant/orders/:id/status`（改状态，仅 {1,2,3,4}，校验归属，无副作用）+ httptest — S
- [ ] **T5** ⚠️ `POST /merchant/orders/:id/redispatch`：仅 `shansong_status∈{-1,60}` → 重新询价(`CalculatePrice`) → 刷新 `ShansongQuoteNo`/清空 `ShansongOrderNo`/状态归 0 → `DispatchShansong` → 回读最新状态 + httptest(mock client) — M
- [ ] **Checkpoint 2** `go test ./...` 全绿；4 接口 curl 验证；确认重新询价费用差异口径；人工 review

## Phase 3 — 前端 SPA（契约稳定后）
- [ ] **T6** `api/order.js` + `views/Orders.vue` 只读运营台（默认「待处理」+ tab 进行中/已完成/全部 + 店铺/日期/类型筛选 + 分页 + 支付/出餐/闪送徽标）+ 路由 + 侧边栏入口 + `npm run build` — M
- [ ] **T7** 行内动作按钮：堂食【出餐】/ 外卖【重新派单】/【改状态】+ 二次确认 + loading 防重 + 成功刷新 — M
- [ ] **Checkpoint 3（端到端）** `go test ./...` 与 `npm run build` 均绿；端到端走查出餐/重新派单/改状态/未支付可见/筛选分页；Ready for review

## 守护测试映射
- T2 → `merchant_order_test.go`（JOIN delivery 字段、含未支付、status/type 筛选、分页、聚合合计）
- T3 → prepare：成功/幂等/越权/不存在
- T4 → status：合法值更新 / 非法值 400 / 越权
- T5 → redispatch：询价成功→派单成功(20) / 派单失败(-1) / 状态非法(400) / 无 client(503) / 越权

## 待确认（开工前）
- 出餐是否回推小程序通知顾客？（默认仅后台，不动 `frontend/`）
- 重新询价新费 > 原配送费 的差额承担 / 是否需二次确认金额？
- `page_size` 默认/上限？单店日订单量级？
- 外卖「卡太久」阈值（待取货 >? 分钟、闪送中 >? 分钟）
