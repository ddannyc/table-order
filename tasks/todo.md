# Todo：订单运营台 — Ship 评审整改 第二轮 (Remediation R2)

详见 `tasks/plan.md`。来源：`/ship` 第二轮 NO-GO（分支 `feat/merchant-order-board`）。
新 Critical：`RedispatchOrder` 缺 `order.Status` 守卫 → 可对已取消订单下付费闪送单（T1/T2 修）。
**git 卫生**：只 stage 该任务文件；禁止 `git add -A`；绝不 stage `frontend/config.js`/`.claude/`。
TDD：先写/改测试再改源码；一任务一提交。

## Phase 1 — GO 阻断（资金安全）
- [x] **T1 ★阻断** `RedispatchOrder` 加 `order.Status==2` 守卫（非已支付→400，配送行不变）+ 测试（status=4&shansong=-1→400）— S
- [ ] **T2 ★阻断** 前端 `canRedispatch` 加 `status===2`；修 `orderBoard.test` 桩（cancelledDelivery→status:4）断言桶互斥/无重派按钮 — S
- [ ] **Checkpoint A（GO）** `go test ./...` + `admin npm test && npm run build` 全绿；已取消单不可重派；ship 可翻 GO

## Phase 2 — 鲁棒性 + CI
- [ ] **T3** limbo 恢复：重派认领纳入 `shansong_status IN (-1,0,60) AND order_no='' AND order.Status==2`；测卡住的 status=0 可重派 — S（依赖 T1）
- [ ] **T4** 新增 `.github/workflows/ci.yml`（Postgres + `REQUIRE_TEST_DB=1` + `CGO_ENABLED=1 go test -race ./...` + admin test/build）+ `setupTestDB` 检查 AutoMigrate 错误 — S
- [ ] **Checkpoint B** CI 在 PR 上真实跑后端(含并发)+前端；人工 review

## Phase 3 — 安全加固（中危跟进）
- [ ] **T5** `UpdateMerchantOrderStatus` 转换白名单（拒 1→2/4→3 等）+ prepare/redispatch/status 审计日志（新模型）+ 测试 — M（依赖 T1）
- [ ] **T6** 入参校验：`shop_id/status` strconv→400、`type` 枚举、检查 Count/Scan/Find `.Error`；+ `page_size>100` 钳制测试 — S
- [ ] **T7** 限流中间件（登录/注册 + redispatch，429，阈值可配）+ 测试 — M
- [ ] **Checkpoint Complete** CI 全绿；ship 再评审 GO；端到端走查（已取消不可重派/限流/审计）；Ready for review

## 守护测试映射
- T1 → redispatch：order.Status=4&shansong=-1 → 400 且行不变；合法(status2)仍成功
- T2 → orderBoard：canRedispatch(status4)=false；inBucket 桶互斥（cancelledDelivery status:4 只进 done）
- T3 → redispatch：status=2&shansong=0&order_no='' → 可重派成功
- T5 → status：非法转换 400 / 合法通过；审计记录写入
- T6 → MerchantOrders：非法 shop_id/status → 400；page_size=500 → 100 条
- T7 → middleware：超阈值 429 / 正常放行

## 本轮不做（验收风险/跟进，见 plan Open Questions）
- 服务端待处理计数（100 行上限漏单）、doRedispatch 取消组件测试、PII 日志脱敏、日期时区一致性
