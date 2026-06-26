# Plan：/ship 评审阻断项修复（外卖/闪送支付链路）

来源：`/ship` 三人评审（code-reviewer / security-auditor / test-engineer）。
范围：仅修复支付/派单链路的阻断项与一项推荐项。业务口径、协议层不变。
基线：绝不 `git add` `frontend/config.js` / `.claude/` / `specs/shansong-delivery.md`；密钥仅 env；一任务一提交，仅 stage 该任务文件 + todo；TDD（RED→GREEN→回归→build→commit）。

## 任务

### BLK1 — 报价密钥 fail-closed + 拒绝负运费（安全 MEDIUM ×2 / 评审 Minor）
**问题**：`delivery.go:34-39` 当 `JWT.Secret` 为空时回退到源码常量 `"dev-quote-secret-fallback"`（可伪造运费令牌）；`order.go:96` 取 `claims.Fee` 无 `>=0` 校验（负运费可少付餐费 / 触发零额自动支付）。
**改动**：`backend/api/handler/delivery.go`
- `quoteSecret()` 改为返回 `([]byte, bool)`，无密钥时 `ok=false`，**不再回退常量**。
- `signQuoteToken` / `verifyQuoteToken` 无密钥时拒绝（签名返回空、验签返回 error）。
- `DeliveryQuote` 在签发前若密钥未配置 → 503「外卖配送暂未开通」。
- `verifyQuoteToken` 增加 `claims.Fee < 0` → error。
**验收/测试**（`delivery_test.go`）：
- 密钥未配置时 `DeliveryQuote` 返回 503（RED）。
- 负运费令牌被 `verifyQuoteToken` 拒绝。
- 既有 token/quote 测试设置 `config.AppConfig.JWT.Secret`（契约对齐）。

### BLK2 — 暴露派单状态：预派单标签 + 派单失败落库（评审 Important ×2 / B1+B2）
**问题**：`shansong_dispatch.go:43-46` 派单失败仅打日志，已付订单被静默搁置；`shansong.go:285-290` 默认状态 0 渲染为「配送中」，掩盖未派单/失败。
**改动**：
- `backend/models/order_delivery.go`：注释明确 `ShansongStatus` 0=待派单、-1=派单失败。
- `backend/services/shansong.go`：`shansongStatusLabels` 增 `0:"待派单"`、`-1:"派单失败"`；未知非零码仍回退「配送中」。
- `backend/services/shansong_dispatch.go`：`CreateOrder` 失败时 best-effort 落 `shansong_status = -1`（const `shansongFailedStatus = -1`），使其可查询、不再被掩盖。（重试/告警列为后续，不在本次范围。）
**验收/测试**（`shansong_dispatch_test.go` / `shansong_test.go`）：
- 派单失败后 `OrderDelivery.ShansongStatus == -1`（RED）。
- `ShansongStatusLabel(0)=="待派单"`、`(-1)=="派单失败"`、未知码仍="配送中"。

### BLK3 — 支付转移原子化（评审 Important / 推荐 R1）
**问题**：`wechatpay_notify.go:73-94` 先读后写、UPDATE 无状态守卫，重复回调可双发 `DistributeReward` + `DispatchShansong`。
**改动**：
- 新增 `markOrderPaidOnce(orderID, paidAt) (bool, error)`：`Where("status = ?", 1).Updates(status=2, paid_at)`，返回 `RowsAffected==1`。
- `WechatPayNotify` 改用该函数；仅当返回 true 才跑奖励分发 + 派单；false（竞态败者）→ 直接 ack SUCCESS。
**验收/测试**：直接测 `markOrderPaidOnce`：status=1 订单两次调用，仅首次返回 true、第二次 false（RED）。

### BLK4 — 回调禁止状态回退（评审 Minor / test-engineer #8 / 推荐 R2）
**问题**：`shansong_notify.go:51-53` 盲写，迟到回调可把 已完成(50) 改回 派单中(20)。
**改动**：`backend/api/handler/shansong_notify.go`：UPDATE 增 `Where("shansong_status NOT IN (50, 60)")`，终态不再被覆盖；`RowsAffected==0`（含已终态）仍返回 200。
**验收/测试**（`shansong_notify_test.go`）：seed 状态 50，投递有效 status=20 回调 → 仍为 50（RED）。

## Checkpoint
- [ ] 三套测试全绿（go / jest 94 / vitest 21）
- [ ] `go build ./...` + `go vet ./...` 通过
- [ ] 仓库无真实密钥；未 add 基线三文件
