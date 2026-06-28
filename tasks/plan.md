# Implementation Plan: 商家后台 · 订单运营台 (Order Action Board)

> Spec: `specs/merchant-admin.md` §1（范围内）、§4（API 契约）、§5 缺口⑥、§9 步骤 1/8。
> Idea: `docs/ideas/order-action-board.md`。

## Overview
为商家后台新增「订单运营台」：默认呈现**待处理队列**（堂食已支付未出餐 →【出餐】；
外卖派单失败/卡太久 →【重新派单】/【改状态】），并支持按店铺/日期/状态/类型筛选、
tab 切换进行中/已完成/全部。未支付订单也展示（带支付状态徽标）。
后端为 Go(Gin) API（Railway），前端为独立 Vue3 SPA（`admin/`，Cloudflare Pages）。

## Architecture Decisions

- **出餐建模 = `Order.PreparedAt *time.Time`，不新增 Status 枚举值。**
  已 grep 确认无处把 `Status==3` 当"已出餐"，加字段非破坏性。`已出餐 = PreparedAt != nil`。
- **契约优先，而非每能力纵切到底。** 后端(Go)与前端(Vue)是两个独立部署物，共享 API 契约。
  按本规划技能建议「共享契约的功能先定契约再并行」：先把列表契约 + 三个动作接口用
  `httptest` 跑绿，再让 SPA 消费稳定契约。避免 3 个动作各自反复改同一个 `Orders.vue`。
- **重新派单必须重新询价。** `services.DispatchShansong` 复用存量 `ShansongQuoteNo` 且当
  `ShansongOrderNo != ""` 时直接 no-op。故 redispatch 接口需：`CalculatePrice`(门店=寄件、
  `OrderDelivery`=收件) → 刷新 `ShansongQuoteNo`、清空 `ShansongOrderNo`、状态归 0 →
  再调 `DispatchShansong`。服务端无需签名 quote token（那只给小程序客户端）。
- **筛选/分页在后端做**；列表的"待处理"判定在前端组合（堂食 `paid && !prepared`，
  外卖 `shansong_status ∈ {-1} 或卡太久`），后端只提供原始字段，避免把易变的运营规则写死后端。

## Dependency Graph

```
T1 Order.PreparedAt (model/migration)
      │
T2 GetMerchantOrders 扩展 (JOIN delivery + 筛选 + 分页 + 含未支付)   ← 列表契约
      ├──────────────┬──────────────┬─────────────────────┐
T3 prepare(出餐)   T4 status(改状态)  T5 redispatch(重新派单)   │  ← 动作接口（可并行）
      └──────────────┴──────────────┴─────────────────────┘
                          │ (契约稳定后)
            T6 前端 Orders.vue 只读三联（列表+筛选+tab+徽标+待处理默认）
                          │
            T7 前端 行内动作按钮（出餐/重新派单/改状态 + 确认 + 刷新）
```

实现顺序自底向上：T1 → T2 →（T3/T4/T5 可并行）→ T6 → T7。

---

## Task List

### Phase 1 — 后端基础（契约）

## Task 1: 给 Order 增加 PreparedAt（出餐时间）

**Description:** 在 `Order` 模型增加可空 `PreparedAt *time.Time`（出餐时间，`已出餐 = 非空`），
确保 AutoMigrate 为既有 `orders` 表追加该列。不新增/改动 Status 枚举。

**Acceptance criteria:**
- [ ] `Order` 含 `PreparedAt *time.Time`，JSON tag `prepared_at`，nullable。
- [ ] 迁移后 `orders` 表存在 `prepared_at` 列（既有行 NULL）。
- [ ] 现有 `go test ./...` 全绿（无回归）。

**Verification:**
- [ ] `cd backend && go build ./... && go test ./...`
- [ ] 本地迁移后确认 `orders` 表含 `prepared_at` 列。

**Dependencies:** None
**Files likely touched:** `backend/models/order.go`（必要时 AutoMigrate 注册处）
**Estimated scope:** XS

---

## Task 2: 扩展 GetMerchantOrders（JOIN 配送明细 + 筛选 + 分页 + 含未支付）

**Description:** 改造 `GET /api/merchant/orders`：LEFT JOIN `order_deliveries` 返回每单配送明细
（`shansong_status` 等），响应增加 `order_type/status/paid_at/prepared_at`；新增 `status`、`type`
过滤与分页（`page`/`page_size`），替换现固定 50 条上限；**不**对未支付(Status=1)做隐式过滤。
`revenue/rewarded` 合计改为按当前筛选条件聚合（非仅当前页）。

**Acceptance criteria:**
- [ ] 响应每单含 `order_type, status, paid_at, prepared_at` 及 `delivery`（外卖单为
      `{recipient_name, recipient_phone, detail_address, delivery_fee, shansong_status, shansong_order_no}`，堂食单为 null）。
- [ ] 支持 `?shop_id=&date=&status=&type=&page=&page_size=`；缺省返回第一页且**含未支付单**。
- [ ] `total` 为筛选后总数，`revenue/rewarded` 为筛选后聚合（不受分页影响）。
- [ ] 仅返回当前 merchant 名下 shop 的订单（沿用现有 shopIDs 约束）。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run MerchantOrders`
- [ ] 手测：构造 1 堂食 + 1 外卖单，`curl` 校验 delivery 字段与分页/筛选。

**Dependencies:** Task 1
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/api/handler/merchant_admin_test.go`（或新增 `merchant_order_test.go`）
**Estimated scope:** M

### Checkpoint: Phase 1（契约就绪）
- [ ] `go test ./...` 全绿；列表响应 schema 冻结（前端据此对接）。
- [ ] Review with human：确认字段名 / 分页参数 / 未支付展示符合预期。

---

### Phase 2 — 后端动作接口（T3/T4/T5 可并行）

## Task 3: 出餐接口 POST /merchant/orders/:id/prepare

**Description:** 新增接口把订单标记为已出餐（置 `PreparedAt = now()`），校验该单所属 shop 归当前 merchant。

**Acceptance criteria:**
- [ ] 成功：`PreparedAt` 置为当前时间，返回 `{message}`；重复调用幂等（已出餐返回成功且不前移时间）。
- [ ] 非本商家订单 → 403/404；订单不存在 → 404。
- [ ] 路由挂在 `merchant` 鉴权组下。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run Prepare`
- [ ] 手测：`curl -X POST .../orders/<id>/prepare` 后列表 `prepared_at` 非空。

**Dependencies:** Task 1, Task 2
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/api/router/router.go`, `*_test.go`
**Estimated scope:** S

---

## Task 4: 改状态接口 PUT /merchant/orders/:id/status

**Description:** 新增接口手动修改 `Order.Status`，校验归属与目标状态合法（仅 {1,2,3,4}）。

**Acceptance criteria:**
- [ ] 请求体 `{status:int}`，仅接受 {1,2,3,4}，非法 → 400。
- [ ] 成功更新返回 `{message}`；非本商家订单 → 403/404。
- [ ] 不触发返利/支付等副作用（纯状态写）。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run OrderStatus`
- [ ] 手测：改状态后列表 `status` 同步。

**Dependencies:** Task 1, Task 2
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/api/router/router.go`, `*_test.go`
**Estimated scope:** S

---

## Task 5: 重新派单接口 POST /merchant/orders/:id/redispatch（含重新询价）⚠️最高风险

**Description:** 对 `派单失败(-1)` 或 `已取消(60)` 的外卖单重新发起闪送：校验归属与状态 →
以门店为寄件、`OrderDelivery` 为收件调 `CalculatePrice` 拿新 quote → 刷新 `ShansongQuoteNo`、
清空 `ShansongOrderNo`、`ShansongStatus=0` → 同步调 `services.DispatchShansong(orderID)` → 回读返回最新状态。

**Acceptance criteria:**
- [ ] 仅当 `ShansongStatus ∈ {-1,60}` 允许；其他 → 400。`services.Shansong==nil` → 503；非本商家 → 403/404。
- [ ] 重新询价成功后写新 `ShansongQuoteNo` 并清空 `ShansongOrderNo`，再派单；
      成功 → `ShansongStatus=20`；失败 → `-1`（不 panic、不阻塞）。
- [ ] 返回体含最新 `shansong_status`（+ 可选新 `delivery_fee`）。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run Redispatch`（mock Shansong client，覆盖成功/询价失败/派单失败/状态非法）。
- [ ] 手测（联调环境）：对一条 -1 单点重新派单，闪送侧出现新运单。

**Dependencies:** Task 1, Task 2
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/api/router/router.go`, `*_test.go`（必要时 `services/shansong_dispatch.go` 暴露同步派单结果）
**Estimated scope:** M

### Checkpoint: Phase 2（动作就绪）
- [ ] `go test ./...` 全绿；4 个接口均 `curl` 验证。
- [ ] 确认 redispatch 费用差异处理口径（见 Open Questions）。
- [ ] Review with human。

---

### Phase 3 — 前端 SPA（消费稳定契约）

## Task 6: Orders.vue 只读运营台（列表 + 筛选 + tab + 徽标 + 待处理默认）

**Description:** 新增 `api/order.js` 与 `views/Orders.vue`，加路由与侧边栏入口。默认 tab「待处理」：
堂食 `已支付 && !已出餐`，外卖 `shansong_status==-1 或卡太久`；另有 进行中/已完成/全部；
支持按当前店铺(`auth.currentShopId`)/日期/类型筛选；行内徽标展示支付/出餐/闪送配送状态。

**Acceptance criteria:**
- [ ] 侧边栏新增「订单」入口，路由 `/orders`（守卫沿用，未登录跳登录）。
- [ ] 默认进入「待处理」，正确聚合两类待办；tab 切换刷新；筛选生效；分页可用。
- [ ] 每行展示：类型、金额、下单时间、支付徽标(未支付/已支付)、堂食出餐徽标、外卖闪送状态文案。
- [ ] `npm run build` 通过，控制台无报错。

**Verification:**
- [ ] `cd admin && npm run build`
- [ ] 手测：登录 → 进入订单 → 待处理/各 tab/筛选/分页 与后端一致，刷新后状态正确。

**Dependencies:** Task 2（及 T3–T5 契约字段）
**Files likely touched:** `admin/src/api/order.js`, `admin/src/views/Orders.vue`, `admin/src/router/index.js`, `admin/src/layouts/AdminLayout.vue`
**Estimated scope:** M

---

## Task 7: 行内动作按钮（出餐 / 重新派单 / 改状态）

**Description:** 在 `Orders.vue` 行内接上三个动作：堂食【出餐】、外卖【重新派单】、通用【改状态】，
带二次确认（重新派单/改状态用 `ElMessageBox.confirm`），成功后 `ElMessage` 提示并刷新当前列表。

**Acceptance criteria:**
- [ ] 堂食「已支付未出餐」行显示【出餐】，点击→确认→出餐徽标变更，行移出待处理。
- [ ] 外卖 `-1/60` 行显示【重新派单】，点击→确认→loading→闪送状态更新；失败有错误提示。
- [ ] 任意行可【改状态】（选 1/2/3/4）；越权/失败由 axios 拦截器统一提示。
- [ ] 动作进行中按钮禁用防重复；`npm run build` 通过。

**Verification:**
- [ ] `cd admin && npm run build`
- [ ] 手测端到端：出餐 / 改状态 / 重新派单 三条路径在联调环境跑通，刷新后状态正确。

**Dependencies:** Task 3, Task 4, Task 5, Task 6
**Files likely touched:** `admin/src/api/order.js`, `admin/src/views/Orders.vue`
**Estimated scope:** M

### Checkpoint: Complete（端到端）
- [ ] 全部验收达成；`go test ./...` 与 `npm run build` 均绿。
- [ ] 端到端走查：堂食出餐、外卖重新派单、改状态、未支付可见、筛选/分页正确。
- [ ] Ready for review。

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| 重新询价费用与顾客已付不一致 | High | MVP 不重新向顾客收费；新费仅记录到 `delivery_fee` 并在返回体/日志暴露差额；口径见 Open Questions，必要时先问商家 |
| `DispatchShansong` 异步 best-effort，同步取结果需小心 | Med | redispatch 内同步调用并回读 `od.ShansongStatus`；不改其在支付流的异步用法 |
| 厨房不在后台实时点【出餐】→ 字段空置、功能形同虚设 | High(产品) | 开工前向真实商家确认出餐履约流程（idea 文档列为头号产品风险） |
| 列表 JOIN + 聚合改动影响既有 `revenue/rewarded` 语义 | Med | 用聚合查询、补 httptest 锁定数值；Checkpoint 1 人工确认口径 |
| 外卖"卡太久"阈值拍脑袋 | Low | 阈值放前端常量、可调；先给保守默认；不写死后端 |

## Open Questions
- 出餐是否需回推小程序通知顾客（"您的餐已出餐"），还是仅后台？（影响是否动 `frontend/`，本计划默认仅后台）
- 重新派单若新报价高于原配送费，差额由谁承担 / 是否需二次确认金额？
- 分页：`page_size` 默认值与上限？预计单店日订单量级？
- 外卖"卡太久"判定阈值（待取货 > ? 分钟、闪送中 > ? 分钟）。
