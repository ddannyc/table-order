# Implementation Plan: 订单运营台 — Ship 评审整改 (Remediation)

> 来源：`/ship` 三专家评审（代码审查 / 安全审计 / 测试工程）对分支 `feat/merchant-order-board` 的结论 **NO-GO**。
> 本计划只做「让决策从 NO-GO → GO」所需的整改 + 高价值测试加固。被三方共同点名的阻断项优先。
> 关联：`specs/merchant-admin.md` §5 缺口⑥、`docs/ideas/order-action-board.md`。

## Overview
修复 ship 评审发现：(1) **阻断项**——`RedispatchOrder` 的「检查-再执行」竞态会重复下闪送付费订单；
(2) 营业额聚合错误地把未支付/已取消计入；(3) 前端「待处理」漏掉可重派的已取消单（闪送 60）。
再补几项被点名的高价值测试缺口与 CI 假性全绿防护。后端 Go(Gin) + 前端 Vue3 SPA(`admin/`)。

## Architecture Decisions

- **重新派单用「原子认领」消除竞态，而非加全局锁。** 先 `CalculatePrice` 询价（窗口外），再用条件
  UPDATE `WHERE id=? AND shansong_status IN (-1,60)` 一次性写入新 quote/清空 order_no/状态归 0，
  仅当 `RowsAffected==1` 才继续派单，否则 409。这样两个并发请求只有一个能认领、只会下一笔闪送单。
  保留原 400 早守卫处理常见的非并发错状态；竞态失败者走 409。
- **营业额口径 = `status IN (2,3)`（已支付/已完成）。** 列表仍展示未支付（产品决定不变），但**金额合计**
  排除未支付(1)与已取消(4)。两者解耦：列表口径 ≠ 金额口径。
- **前端 triage 规则单一可信源。** 把 `Orders.vue` 的 `inBucket` 抽进 `orderBoard.js` 与 `needsAction`
  并列，并修正 `needsAction` 纳入闪送 `60`（与 `canRedispatch` 对齐），消除「有重派按钮却不在待处理」的矛盾。
- **CI 不再容忍跳过。** 后端 handler 测试依赖本地 Postgres，缺库会 `t.Skip` 导致假性全绿；以环境变量
  `REQUIRE_TEST_DB` 兜底：设置时缺库直接 `Fatal`，本地默认仍 skip。不改既有本地开发体验。
- **本次不做（已记录为验收风险/跟进）：** 限流、状态机+审计日志、服务端「待处理」计数。见 Open Questions。
  （JWT 锁算法、prepare 后端不变量已拉入本批：见 T7/T8。）

## Dependency Graph

```
独立可并行，但按风险排序（先 fail-fast 修资金安全）：

T1 RedispatchOrder 原子认领（资金安全）★阻断
T2 营业额聚合限定 status IN (2,3)
        └── 后端契约稳定 → Checkpoint A（GO 的后端门槛）
T3 needsAction 纳入 60 + 抽出 inBucket（前端 triage 修正）
        └── Checkpoint B（GO 门槛达成）
T4 列表测试缺口：shop_id / date / page_size 上限（纯补测）
T5 redispatch 测试缺口：询价失败 502 / 非外卖单 400（需 mock 支持错误）
T6 CI：缺 Postgres 时 handler 测试判失败（REQUIRE_TEST_DB）
```

实现顺序：T1 → T2 →（Checkpoint A）→ T3 →（Checkpoint B / GO）→ T4 → T5 → T6。

---

## Task List

### Phase 1 — 资金安全 + 金额正确性（GO 的后端门槛）

## Task 1: RedispatchOrder 原子认领，消除重复派单竞态 ★阻断

**Description:** 重排 `RedispatchOrder`：先询价（`CalculatePrice`），再以条件 UPDATE 认领该配送行
（`WHERE id=? AND shansong_status IN (-1,60)` 写入新 `shansong_quote_no`、清空 `shansong_order_no`、
`shansong_status=0`、更新 `delivery_fee`），仅当 `RowsAffected==1` 继续调用 `DispatchShansong`，
否则返回 409。保留早期 400 守卫处理非并发错状态。

**Acceptance criteria:**
- [ ] 并发两个 redispatch（goroutine + 计数 mock）：仅 1 个 200、其余 409/400；`orderPlace` 仅被调用一次。
- [ ] 询价在认领之前：`CalculatePrice` 失败时返回 502 且**配送行未被改动**（quote/order_no 保持原值）。
- [ ] 单次成功路径仍：刷新 quote、清空旧 order_no、派单成功后 `shansong_status=20` 并返回最新状态。
- [ ] 新增 redispatch 的跨商家归属测试（403/404）。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run Redispatch -race`
- [ ] 手测（联调）：对一条 -1 单连点两次，闪送侧只出现一笔新运单。

**Dependencies:** None（最高风险，先做）
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/api/handler/merchant_order_test.go`
**Estimated scope:** M

---

## Task 2: 营业额/返利聚合限定为 status IN (2,3)

**Description:** `GetMerchantOrders` 的 revenue/rewarded 聚合加状态过滤，仅统计已支付(2)/已完成(3)；
列表本身仍返回含未支付/已取消的全集（不变）。

**Acceptance criteria:**
- [ ] 含 1 笔未支付(1) + 1 笔已取消(4) + 2 笔已支付(2) 时，`revenue` 只累计已支付/已完成。
- [ ] `total`（列表条数）仍包含未支付/已取消（列表口径不变）。
- [ ] 既有分页/聚合测试不回归。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run MerchantOrders`

**Dependencies:** None
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/api/handler/merchant_order_test.go`
**Estimated scope:** S

### Checkpoint A — 后端门槛
- [ ] `go test ./... -race` 全绿；阻断项闭合。
- [ ] Review with human：确认营业额口径（2,3）与 409 语义符合预期。

---

### Phase 2 — 前端 triage 正确性（GO 门槛达成）

## Task 3: needsAction 纳入闪送 60 + 抽出 inBucket 至 orderBoard.js

**Description:** 修正 `needsAction`：外卖单 `shansong_status ∈ {-1,60}` 均为待处理（与 `canRedispatch`
对齐），消除「显示重派按钮却不在待处理队列」的矛盾。把 `Orders.vue` 的 `inBucket` 分桶逻辑抽到
`orderBoard.js` 成纯函数并补测；`Orders.vue` 改为引用它。

**Acceptance criteria:**
- [ ] `needsAction` 对 `{order_type:'delivery', delivery:{shansong_status:60}}` 返回 true；既有用例不回归。
- [ ] `inBucket(order,'pending'|'active'|'done'|'all')` 抽为导出纯函数并有单测覆盖四个桶。
- [ ] 已取消外卖(60)单出现在「待处理」桶，不再落入「进行中」。
- [ ] `npm test` 全绿、`npm run build` 通过。

**Verification:**
- [ ] `cd admin && npm test -- orderBoard && npm run build`
- [ ] 手测：一条 shansong=60 的外卖单出现在「待处理」并可重新派单。

**Dependencies:** None（前端独立）
**Files likely touched:** `admin/src/utils/orderBoard.js`, `admin/src/utils/orderBoard.test.js`, `admin/src/views/Orders.vue`
**Estimated scope:** S

### Checkpoint B — GO 门槛达成
- [ ] 阻断项(T1) + 两项建议修复(T2,T3) 均完成、测试绿、build 绿。
- [ ] 可将 ship 决策翻为 GO（剩余为加固项，可作快速跟进）。
- [ ] Review with human。

---

### Phase 3 — 测试加固（提高信心，可快速跟进）

## Task 4: 补列表端点测试缺口（shop_id / date / page_size 上限）

**Description:** 纯补测，不改源码：覆盖真实生产路径 `?shop_id=`（多店商家筛单）、`?date=` 过滤与非法
日期被静默忽略的行为、`page_size>100` 被钳制到 100、空店铺商家返回 200+空列表。

**Acceptance criteria:**
- [ ] `?shop_id=` 只返回该店订单；跨店不泄漏。
- [ ] `?date=YYYY-MM-DD` 命中当日；非法 `?date=foo` 退化为不过滤（与现实现一致）。
- [ ] `?page_size=999` 实际返回 ≤100；空店铺商家 → 200 且 `orders:[]`。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run MerchantOrders`

**Dependencies:** None
**Files likely touched:** `backend/api/handler/merchant_order_test.go`
**Estimated scope:** S

---

## Task 5: 补 redispatch 测试缺口（询价失败 502 / 非外卖单 400）

**Description:** 扩展 `shansongMock` 支持「询价返回错误信封/非200」，覆盖 `CalculatePrice` 失败 → 502
且配送行未改动；以及对无 `OrderDelivery` 的堂食单调用 redispatch → 400。

**Acceptance criteria:**
- [ ] 询价失败 → 502，且 `shansong_quote_no/order_no` 保持原值（验证「先询价后认领」不留脏数据）。
- [ ] 堂食单（无配送行）→ 400 "not a delivery order"。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run Redispatch`

**Dependencies:** Task 1（认领顺序确定后再补这些断言）
**Files likely touched:** `backend/api/handler/merchant_order_test.go`
**Estimated scope:** S

---

### Phase 4 — CI 防护

## Task 6: CI 缺 Postgres 时 handler 测试判失败（不再假性全绿）

**Description:** `setupTestDB` 当前在连不上库时 `t.Skip`，导致 CI 无库时 `go test` 假绿。以环境变量
`REQUIRE_TEST_DB` 兜底：设置该变量且连库失败时 `t.Fatal`；未设置时维持 `t.Skip`（本地不受影响）。
CI 配置设 `REQUIRE_TEST_DB=1` 并提供 `table_order_test`。

**Acceptance criteria:**
- [ ] 未设 `REQUIRE_TEST_DB`：无库时仍 skip（本地行为不变）。
- [ ] 设 `REQUIRE_TEST_DB=1` 且无库：测试 Fatal（CI 会失败）。
- [ ] 有库时两种情况都正常跑过。

**Verification:**
- [ ] 本地：`go test ./api/handler/`（skip 或过）；`REQUIRE_TEST_DB=1 go test ./api/handler/`（有库则过）。
- [ ] CI 配置项随附说明（若有 CI 配置文件则一并更新）。

**Dependencies:** None
**Files likely touched:** `backend/api/handler/handler_test.go`（+ 可能的 CI 配置）
**Estimated scope:** S

## Task 7: PrepareOrder 拒绝非已支付单（后端不变量）

**Description:** 后端不再依赖前端守卫：`PrepareOrder` 对未支付(1)或已取消(4)的订单返回 400，
仅允许已支付(2)/已完成(3)标记出餐。保持已支付单的幂等行为。

**Acceptance criteria:**
- [ ] 未支付(1)单 prepare → 400 且 `prepared_at` 仍为空；已取消(4)单 → 400。
- [ ] 已支付(2)单 prepare → 200 且置 `prepared_at`；重复仍幂等。
- [ ] 既有 prepare 用例不回归。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run Prepare`

**Dependencies:** None（与 T1/T2 同属后端，建议紧随）
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/api/handler/merchant_order_test.go`
**Estimated scope:** S

---

## Task 8: JWT 锁定签名算法 HS256（防御加固）

**Description:** `ParseWithClaims` 增加 `jwt.WithValidMethods([]string{"HS256"})`，杜绝算法混淆/`alg:none`
的未来风险。全仓库鉴权（含本批新接口）依赖此中间件。

**Acceptance criteria:**
- [ ] 解析仅接受 HS256；非 HS256/`none` 签名的 token 被拒（401）。
- [ ] 既有鉴权相关测试与 `go test ./...` 不回归。

**Verification:**
- [ ] `cd backend && go test ./...`
- [ ] 若有 middleware 测试：`go test ./middleware/`

**Dependencies:** None（独立，放最后做）
**Files likely touched:** `backend/middleware/auth.go`（+ 可能的 `middleware/*_test.go`）
**Estimated scope:** XS

---

### Checkpoint: Complete
- [ ] `go test ./... -race` 与 `cd admin && npm test && npm run build` 全绿。
- [ ] ship 决策可翻 GO；端到端联调走查（含并发重派只出一单）。
- [ ] Ready for review / 合并。

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| 并发测试 flaky | Med | 用条件 UPDATE 使结果确定（认领唯一）；计数 mock 加锁；`-race` 跑 |
| 改 revenue 口径影响看板既有展示 | Med | 仅动 orders 列表聚合；看板 dashboard/stats 是另接口不动；补测锁数值 |
| 抽 inBucket 改动 Orders.vue 触发回归 | Low | 纯逻辑搬迁 + 单测覆盖；build + 手测 |
| 改 handler_test skip 行为影响本地 | Low | 默认仍 skip，仅 `REQUIRE_TEST_DB` 时 Fatal |

## Open Questions（本次不做，待定/跟进）
- 服务端「待处理」计数/筛选（解决 100 行上限漏单）——量起来前是否需要？
- `redispatch` 与登录接口限流（安全中危）——单独加固任务？
- `UpdateMerchantOrderStatus` 状态机 + 审计日志——是否下一批纳入？
- 重新询价新费 > 原费 是否需前端展示差额/二次确认？（当前仅记录不补收，已定）

> 已拉入本批：JWT 锁 HS256（T8）、PrepareOrder 拒绝非已支付（T7）。
