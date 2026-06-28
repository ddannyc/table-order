# TODO — 订单运营台 Ship Remediation R3

分支 `feat/merchant-order-board`；勿动/勿 stage `frontend/config.js`、`.claude/`；逐任务单独提交。

## Phase 1 — GO 阻断（安全）
- [x] **T1 ★阻断(HIGH)** 可信代理配置：`ServerConfig.TrustedProxies` + `TRUSTED_PROXIES` 环境变量（仿 `AllowedOrigins`）；`main.go` 调 `r.SetTrustedProxies(...)`，空列表则告警；测：伪造 `X-Forwarded-For` 不能换桶（仍 429）+ config 解析 — M
- [x] **T2 ★阻断(MED/资金)** 重派原子认领纳入 `order.Status=2`（关联子查询 `order_id IN (SELECT id FROM orders WHERE id=? AND status=2)`），`RowsAffected==0→409`；测：报价窗口内并发改 `status=4` → 0 次 orderPlace（`quoteDelay` 竞态，仿 ConcurrentDispatchesOnce）— M
- [x] **Checkpoint A（GO）** `go test ./...` + admin `npm test`/`build` 全绿；伪造 XFF 不能重置限流；中途取消单不可派单；两个阻断可翻 resolved

## Phase 2 — 鲁棒性（建议修）
- [x] **T3** 限流器加固：`allow()` 清空键 `delete`；测窗口复位（注入 `now`，零 sleep）、键淘汰、`ByUserID` 分桶+无 user 回退 IP — S
- [x] **T4** 审计写失败可观测：`logOrderAction` 捕获并 `log.Printf`（保持提交后审计，不回滚动作）+ 注释说明 — S

## Phase 3 — 测试覆盖
- [ ] **T5** handler 覆盖缺口（依赖 T2）：limbo `shansong_status=0`→重派→400 表征测；重派成功写 `action="redispatch"` 审计断言；状态机白名单表测（合法 1→4/2→3/2→4 通过；非法 2→1/2→2/3→1/3→4/4→2/4→3→400 且状态不变）— S/M
- [ ] **Checkpoint Complete** `go test ./...`（Redispatch `-count=5`）+ admin 全绿；重跑 `/ship` 期望 GO（仅剩已知风险）；端到端走查

## 已知风险 / 本轮不做
- 限流器仅单实例（多 pod 需 Redis）— 文档化
- limbo `status=0` 基于时间的对账清扫 — 仅表征当前 400，不改行为
- 预存 `CreateMerchantOrder` 用客户端 `amount` 建 `Status:2` 单（在 main，超本支范围）— 另议
- `TRUSTED_PROXIES` 具体 Railway 边缘 CIDR — 部署时设值（Open Question）
