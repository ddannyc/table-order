# Implementation Plan: 订单运营台 — Ship 评审整改 第二轮 (Remediation R2)

> 来源：`/ship` 第二轮评审（NO-GO）。第一轮阻断项已闭合并验证；本轮修复**新发现的 Critical**
> （由 T4 改状态 × T5 重新派单 交互引入）+ 鲁棒性/CI/安全跟进。
> 关联：`specs/merchant-admin.md` §5 缺口⑥、`tasks/plan.archived-ship-remediation.md`（第一轮）。

## Overview
核心阻断：`RedispatchOrder` 缺 `order.Status` 守卫——商家用 改状态 把订单置为已取消(4)后，配送行仍是
`shansong_status ∈ {-1,60}`，重新派单会对**已取消订单下一笔付费闪送单**。这是第一轮 T7（prepare 不信任
前端门）在重新派单路径上的遗漏。修复 = 后端要求 `order.Status==2` + 前端 `canRedispatch` 同步要求
`status===2`（顺带修桶互斥与测试桩）。再补：limbo 状态恢复、真正生效的 CI、状态机+审计、入参校验、限流。

## Architecture Decisions

- **重新派单严格要求 `order.Status == 2`（已支付）。** 失败重派(-1)与闪送侧取消(60)在订单仍已支付时合法；
  商家自己取消(4)/未支付(1)/已完成(3) 一律不可重派。前端 `canRedispatch` 同步加 `status===2`，使
  `needsAction`(待处理) 与 `done`(status 3/4) 桶恢复互斥，且已取消单不再显示重派按钮。
- **limbo `status=0` 纳入可重派集（带 `order_no==''` 守卫）。** 认领把 `shansong_status` 先置 0 再派单；
  崩溃会卡在 0。让重派认领匹配 `shansong_status IN (-1,0,60) AND shansong_order_no='' AND order.Status=2`，
  使被卡订单可由商家自助重派。正常支付流首次派单近乎与支付同刻发生且前端不为 0 显示按钮，残余竞态可忽略（文档化）。
- **CI 真正执行测试。** 新增 `.github/workflows/ci.yml`：起 Postgres、设 `REQUIRE_TEST_DB=1`、`CGO_ENABLED=1`
  跑 `go test -race ./...` + admin `npm ci/test/build`。并在 `setupTestDB` 检查 `AutoMigrate` 错误。
  （此前 plan/todo 写"已跑 -race"不准确：cgo 关闭，从未真正跑过。）
- **状态机 + 审计为下一阶段安全加固**，非 GO 门槛；GO 仅取决于 Phase 1。

## Dependency Graph

```
T1 后端 redispatch 要求 order.Status==2 ★阻断
T2 前端 canRedispatch 要求 status===2 + 修桶/测试桩 ★阻断
        └── Checkpoint A（GO 门槛）
T3 limbo: 可重派集纳入 0（带 order_no='' 守卫）   [依赖 T1 的 status 守卫]
T4 CI 流水线（让 REQUIRE_TEST_DB 生效 + -race）+ AutoMigrate 错误检查
        └── Checkpoint B
T5 UpdateMerchantOrderStatus 状态机 + 订单操作审计日志
T6 入参校验（shop_id/status→400、type 枚举、检查被忽略的 GORM error）+ page_size 上限测试
T7 限流中间件（登录/注册 + redispatch）
        └── Checkpoint Complete
```

实现顺序：T1 → T2 →（Checkpoint A / GO）→ T3 → T4 →（Checkpoint B）→ T5 → T6 → T7。

---

## Task List

### Phase 1 — GO 阻断（资金安全）

## Task 1: RedispatchOrder 要求 order.Status==2 ★阻断

**Description:** 在 `RedispatchOrder` 的 `loadOwnedOrder` 之后、询价之前，加 `order.Status==2` 守卫；
非已支付（未支付1/已完成3/已取消4）返回 400，杜绝对已取消订单下付费闪送单。镜像第一轮 T7。

**Acceptance criteria:**
- [ ] `order.Status != 2` → 400，且配送行未变动（`shansong_quote_no/order_no/status` 原值）。
- [ ] 已支付(2)且 `shansong_status∈{-1,60}` 的合法重派仍成功（既有用例不回归）。
- [ ] 新增测试：order.Status=4 且 shansong_status=-1 → 400（镜像 `TestPrepareOrder_RejectsCancelled`）。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run Redispatch`

**Dependencies:** None（最高风险，先做）
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/api/handler/merchant_order_test.go`
**Estimated scope:** S

---

## Task 2: 前端 canRedispatch 要求 status===2 + 修桶互斥与测试桩 ★阻断

**Description:** `canRedispatch` 增加 `order.status === 2`，使已取消(4)外卖单不再同时落入「待处理」与
「已完成」两个桶、也不再显示重派按钮。修正 `orderBoard.test.js` 中名为 cancelledDelivery 实为 `status:2`
的桩，改为真正的 `status:4`，并断言其只进 done 桶、不显示重派。

**Acceptance criteria:**
- [ ] `canRedispatch({status:4, delivery:{shansong_status:-1}})` → false；`{status:2, shansong:-1/60}` → true。
- [ ] 测试桩 cancelledDelivery 改为 `status:4`，断言 `inBucket(...,'pending')===false` 且 `'done'===true`（桶互斥）。
- [ ] `npm test` 全绿、`npm run build` 通过。

**Verification:**
- [ ] `cd admin && npm test -- orderBoard && npm run build`

**Dependencies:** None（前端独立；与 T1 同属一个 GO 门槛）
**Files likely touched:** `admin/src/utils/orderBoard.js`, `admin/src/utils/orderBoard.test.js`
**Estimated scope:** S

### Checkpoint A — GO 门槛
- [ ] `go test ./...` 全绿、`admin npm test && npm run build` 全绿。
- [ ] 新 Critical 闭合：已取消订单不可重派（后端 400 + 前端无按钮/桶互斥）。
- [ ] Review with human → ship 可翻 GO。

---

### Phase 2 — 鲁棒性 + CI

## Task 3: 可恢复的 limbo（重派认领纳入 shansong_status=0）

**Description:** 认领把状态先置 0 再同步派单；崩溃会卡在 0（不在 {-1,60}，无法再重派）。把重派认领的
条件 UPDATE 改为匹配 `shansong_status IN (-1,0,60) AND shansong_order_no='' `（叠加 T1 的 `order.Status==2`），
使卡住的已支付订单可被商家自助重派。文档化正常支付流首派的极小残余窗口。

**Acceptance criteria:**
- [ ] 构造 `order.Status=2, shansong_status=0, shansong_order_no=''` 的订单：重派成功（200，最终 20）。
- [ ] 仍拒绝 `order_no!=''`（已派单）与 `order.Status!=2`。
- [ ] 既有并发/成功/失败/非法用例不回归。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run Redispatch`

**Dependencies:** Task 1
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/api/handler/merchant_order_test.go`
**Estimated scope:** S

---

## Task 4: CI 流水线让 REQUIRE_TEST_DB 真正生效 + AutoMigrate 错误检查

**Description:** 新增 `.github/workflows/ci.yml`：Postgres service、`REQUIRE_TEST_DB=1`、`CGO_ENABLED=1`
跑 `go test -race ./...`；admin `npm ci && npm test && npm run build`。在 `setupTestDB` 检查并 fail
on `AutoMigrate` 错误（当前忽略）。消除"假性全绿"与不准确的 -race 声明。

**Acceptance criteria:**
- [ ] CI 工作流存在：起 Postgres、设 `REQUIRE_TEST_DB=1`、`CGO_ENABLED=1 go test -race ./...`、admin 测试+构建。
- [ ] `setupTestDB` 对 `AutoMigrate` 返回错误 `t.Fatalf`（不再静默）。
- [ ] 本地 `go test ./...`（无 env）仍可 skip；有库时正常通过。

**Verification:**
- [ ] 本地 `REQUIRE_TEST_DB=1 go test ./api/handler/`（有库则过）。
- [ ] CI 配置语法自检（actionlint 或人工核对）。

**Dependencies:** None
**Files likely touched:** `.github/workflows/ci.yml`（新增）, `backend/api/handler/handler_test.go`
**Estimated scope:** S

### Checkpoint B — 鲁棒性/CI
- [ ] CI 在 PR 上真实跑起后端（含并发测试）+ 前端。
- [ ] Review with human。

---

### Phase 3 — 安全加固（中危跟进）

## Task 5: 订单状态转换状态机 + 操作审计日志

**Description:** `UpdateMerchantOrderStatus` 加合法转换白名单（如仅 `2→3`、`2→4`、`1→4`；拒绝复活已取消/
凭空标记已支付）。为 prepare/redispatch/status 三类商家操作写审计记录（actor merchant_id、order_id、
old→new、时间）。消除「自助把未支付标记为已支付以虚增营业额」与无留痕问题。

**Acceptance criteria:**
- [ ] 非法转换（如 `1→2`、`4→3`、`3→1`）→ 400；合法转换通过。
- [ ] 每次 prepare/redispatch/status 成功写一条审计记录（含 actor/old→new/时间）。
- [ ] 新增模型/表的迁移与单测；`go test ./...` 全绿。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run 'Status|Audit'`

**Dependencies:** Task 1（redispatch 守卫先就位）
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/models/`（新增审计模型）, `backend/api/handler/handler_test.go`, `*_test.go`
**Estimated scope:** M

---

## Task 6: 入参校验 + 被忽略的 DB 错误 + page_size 上限测试

**Description:** `GetMerchantOrders` 用 `strconv.Atoi` 解析 `shop_id`/`status`（失败 400），校验 `type` 枚举；
检查 `Count/Scan/Find` 的 `.Error`（记录或 500）。补 `page_size>100` 钳制到 100 的测试。

**Acceptance criteria:**
- [ ] `?status=abc`/`?shop_id=xx` → 400（非静默空结果）；`?type=bogus` → 400 或忽略（择一并测）。
- [ ] `page_size=500` + >100 行：返回正好 100 条（上限分支被覆盖）。
- [ ] 既有列表用例不回归。

**Verification:**
- [ ] `cd backend && go test ./api/handler/ -run MerchantOrders`

**Dependencies:** None
**Files likely touched:** `backend/api/handler/merchant_order.go`, `backend/api/handler/merchant_order_test.go`
**Estimated scope:** S

---

## Task 7: 限流中间件（登录/注册 + redispatch）

**Description:** 新增 per-IP + per-account 限流中间件，挂在 `/api/auth/login`、`/api/merchant/login`、
`/api/merchant/register`（防爆破）与 `/api/merchant/orders/:id/redispatch`（防配额/费用滥用）。

**Acceptance criteria:**
- [ ] 超过阈值的连续请求返回 429。
- [ ] 正常频率不受影响；单测覆盖放行/限流两路径。
- [ ] 阈值可配置（config/env）。

**Verification:**
- [ ] `cd backend && go test ./middleware/`

**Dependencies:** None
**Files likely touched:** `backend/middleware/ratelimit.go`（新增）, `backend/api/router/router.go`, `backend/middleware/*_test.go`
**Estimated scope:** M

### Checkpoint: Complete
- [ ] `go test ./... -race`（CI）+ `admin npm test && npm run build` 全绿。
- [ ] ship 再评审可 GO；端到端走查（已取消单不可重派、限流生效、审计有记录）。
- [ ] Ready for review / 合并。

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| limbo 纳入 0 与正常支付流首派竞态 | Low | 叠加 `order_no=''` + `order.Status==2`；前端不为 0 显按钮；窗口极小且文档化 |
| 状态机白名单与现有手动流冲突 | Med | 白名单先宽（仅禁危险转换），保留 prepare/redispatch 路径；充分单测 |
| 限流误伤正常商家 | Med | 阈值保守可配；仅登录/注册/redispatch，不动读接口 |
| CI 首次接入环境差异 | Low | 用官方 postgres service container；与本地 DSN 对齐 |

## Open Questions（待定/跟进，本轮可不做）
- 服务端「待处理」计数（解决 100 行上限漏单）——量起来前是否需要？
- `doRedispatch` 取消确认不调 API 的组件测试（前端 SFC 唯一有后果的逻辑）。
- 列表 PII（recipient_*）的日志脱敏与前端缓存策略——合规确认。
- 日期过滤的 DB/应用时区一致性（跨时区临界）。
