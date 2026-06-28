# Todo：订单运营台 — Ship 评审整改 (Remediation)

详见 `tasks/plan.md`。来源：`/ship` 评审结论 NO-GO（分支 `feat/merchant-order-board`）。
目标：闭合阻断项 + 两项建议修复 → 翻 GO；再补高价值测试与 CI 防护。
**git 卫生**：只 stage 该任务文件；禁止 `git add -A`；绝不 stage `frontend/config.js`/`.claude/`。
TDD：先写/改测试再改源码；一任务一提交。

## Phase 1 — 资金安全 + 金额正确性（GO 后端门槛）
- [x] **T1 ★阻断** `RedispatchOrder` 原子认领：先 `CalculatePrice` → 条件 UPDATE `WHERE shansong_status IN (-1,60)`（`RowsAffected==1` 才派单，否则 409）→ `DispatchShansong`。并发测试(仅一单/orderPlace 一次) + 跨商家测试 + 询价失败不留脏数据 — M
- [x] **T2** revenue/rewarded 聚合限定 `status IN (2,3)`（列表仍含未支付/已取消）+ 测试（混合状态断言金额排除 1/4）— S
- [x] **Checkpoint A** `go test ./... -race` 全绿；阻断闭合；人工确认金额口径与 409 语义

## Phase 2 — 前端 triage 正确性（GO 门槛达成）
- [x] **T3** `needsAction` 纳入闪送 `60`（与 `canRedispatch` 对齐）+ 抽 `inBucket` 至 `orderBoard.js` 并补测 + `Orders.vue` 改引用 — S
- [x] **Checkpoint B（GO）** T1+T2+T3 完成、`npm test`/`npm run build` 绿；ship 可翻 GO

## Phase 3 — 测试加固（快速跟进）
- [x] **T4** 补列表测试：`shop_id` 筛选(生产路径)/`date`+非法日期静默忽略/`page_size>100` 钳制/空店铺 200 空列表 — S
- [x] **T5** 补 redispatch 测试：扩展 mock 支持询价错误 → 502 且不留脏数据；非外卖单 → 400（依赖 T1）— S

## Phase 4 — CI 防护 + 安全加固
- [x] **T6** `setupTestDB`：`REQUIRE_TEST_DB` 设置且无库时 `t.Fatal`（默认仍 skip，本地不变）；CI 设该变量并提供 `table_order_test` — S
- [ ] **T7** `PrepareOrder` 拒绝未支付(1)/已取消(4) → 400（仅 2/3 可出餐，保持幂等）+ 测试 — S
- [ ] **T8** JWT `ParseWithClaims` 加 `WithValidMethods(["HS256"])`（拒非 HS256/none）+ 不回归 — XS
- [ ] **Checkpoint Complete** `go test ./... -race` + `admin npm test && npm run build` 全绿；端到端走查（并发重派只出一单）；Ready for review/合并

## 守护测试映射
- T1 → redispatch：并发(goroutine+计数 mock)/跨商家/询价失败不留脏；强化既有 reject 用例断言状态未变
- T2 → MerchantOrders：混合 1/2/3/4 断言 revenue 仅 2+3
- T3 → orderBoard：needsAction(60)=true；inBucket 四桶；Orders.vue 引用
- T4 → MerchantOrders：shop_id/date/非法 date/page_size 钳制/空店铺
- T5 → redispatch：quote-fail 502 + 不留脏；dine-in 400

## 本次不做（验收风险/跟进，见 plan Open Questions）
- 服务端待处理计数（100 行上限漏单）、限流、状态机+审计日志
- （JWT 锁 HS256 → T8；prepare 后端不变量 → T7，已拉入本批）
