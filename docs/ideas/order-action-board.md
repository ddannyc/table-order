# 商家后台 · 订单运营台 (Order Action Board)

> 状态：已定方向，待排期实现。配套 spec：`specs/merchant-admin.md`（本模块原列为 v1 范围外，现解冻）。

## Problem Statement
How might we give a merchant one place to see every order that needs their
attention — 堂食 waiting to be 出餐, 外卖 that failed/stalled in 闪送 dispatch —
and act on it, without making them babysit a dashboard?

## Recommended Direction
A single "订单运营台" that defaults to a **待处理 (needs-action) queue**, not a
log. Two order types unify as "orders waiting on the merchant":
 - 堂食: 已支付未出餐 → 一键【出餐】
 - 外卖: 派单失败 / 卡太久 → 【重新派单】/【改状态】
Tabs for 进行中 / 已完成 / 全部 + 按店铺/日期筛选。未支付订单也展示（商家需知道
"是否已支付"），用支付状态徽标区分。The happy path fades; the few things that
need you float to the top. Picked over a kanban (re-adds babysitting) and a
flat全量 table (scan fatigue).

## Decisions (locked)
- **出餐 建模**：新增 `Order.PreparedAt *time.Time`，**不**新增 Status 枚举值
  （非破坏性，`已出餐` = `PreparedAt != null`）。
- **未支付订单**：展示（不隐藏），用支付状态区分 `Status=1 未支付` vs `>=2 已支付`。

## Key Assumptions to Validate
- [ ] 加 `PreparedAt` 前 grep 所有 `Status` 读取点，确认无处把 `Status==3`
      当作 "已出餐" 语义（迁移前核对）。
- [ ] 厨房会实时在后台点【出餐】（否则字段形同虚设）——开工前向真实商家确认
      出餐履约流程。最大的"产品"风险，不是技术风险。
- [ ] `services.DispatchShansong` 在重新派单时会**重新询价**（旧
      `ShansongQuoteNo` 可能已失效）——接按钮前先读服务实现。

## MVP Scope
IN:
 - Backend: 扩展 `GetMerchantOrders` —— LEFT JOIN `order_deliveries`，返回
   `order_type / status / paid_at / prepared_at` + `delivery{shansong_status,...}`；
   增加 status/type 筛选 + 分页（或处理 50 条上限）。
 - Backend: `Order` 加 `PreparedAt *time.Time`（迁移）。
 - 3 个接口（均校验 shop 归属当前 merchant）：
     POST /api/merchant/orders/:id/prepare      → 置 PreparedAt（出餐）
     POST /api/merchant/orders/:id/redispatch   → services.DispatchShansong（仅限失败/已取消）
     PUT  /api/merchant/orders/:id/status       → 手动改状态
 - admin/: `api/order.js` + `views/Orders.vue`（默认"待处理" + tab + 筛选
   + 行内动作）+ 路由 + 侧边栏入口。
 - 3 个新 handler 的 Go httptest（沿用 spec §7 风格）。
OUT（本期不做）：
 - Kanban / 拖拽，推送 & 短信派单失败告警，闪送自动重试自愈，
   退款/refund，图表，堂食返利明细报表。

## Not Doing (and Why)
- 推送/短信派单失败告警 —— 强大，但改动通知行为，超出"补拼图"；等运营台跑通
  工作流后再议。
- 闪送失败自动重试自愈 —— 静默自动产生配送费有风险；v1 保持重新派单为人工决策。
- 退款/refund —— 未勾选；涉及资金流动，应单列范围。
- 堂食/外卖拆两个独立视图 —— 过早拆分；先合并，某一侧体量起来再拆。

## Open Questions
- 出餐是否需要回推小程序通知顾客（"您的餐已出餐"），还是仅后台？
- 分页 vs 仅提高 50 条上限 —— 预计单店日订单量？
- 外卖"卡太久"阈值 —— 各状态真实分钟数（待取货>?、闪送中>?）。
