# 实施计划：/ship 评审整改（双模式下单上线前修复）

## 概述
`feat/dual-mode-ordering` 分支 /ship 评审结论为 **NO-GO**。本计划把评审发现的问题切成可独立验证的纵向任务，按风险分级。执行后该分支可翻为 **GO**。

所有修复遵循项目 TDD 约定（RED → GREEN → 回归 → 一任务一提交）。后端 `go test ./...`（需本地 Postgres），前端 `node node_modules/jest/bin/jest.js`，admin `node node_modules/vitest/vitest.mjs run`。

**基线纪律：** 绝不 `git add` `frontend/config.js`、`.claude/`、`specs/shansong-delivery.md`；每个提交只暂存该任务相关文件。

## 分级
- **P0 阻断项**：必须修，否则不可上线。
- **P1 资金路径风险**：与钱直接相关，强烈建议随本次一起修（修完即干净 GO）。
- **P2 正确性/信息泄露**：低风险但应修；可并入本次或下次。

## 依赖关系
- T1–T4 都改 `backend/api/handler/order.go` / `shop.go`，互不冲突但同文件，按顺序提交避免 rebase 噪音。
- T5（前端 wx:key）、T6（admin/后端规格）独立，可任意顺序。
- 建议执行序：T1 → T2 → T3 → T4 → T5 → T6 → Checkpoint。

---

## 任务列表

### Task 1 — 【P0 阻断】规格必选守卫，堵住 spec_id:0 低价下单
**问题：** `order.go:86-102` 当 `item.SpecID == 0` 时价格回退到 `product.Price`，但未校验该商品是否本就有规格。客户端对「有规格」的商品传 `spec_id:0`，即可按商品基础价（可能远低于规格价）下单。
**做法：** 当 `item.SpecID == 0` 时，查询该商品是否存在 `status=1` 的规格；若存在则返回 400「请选择规格」。`single-default-spec`（无规格）商品不受影响。
**Acceptance:**
- [ ] 对「有上架规格」的商品传 `spec_id:0` → 400，订单不创建。
- [ ] 无规格商品传 `spec_id:0` → 仍按 `product.Price` 正常下单（行为不变）。
- [ ] 传合法 `spec_id` → 按规格价下单（行为不变）。
**Verification:** [ ] `backend/api/handler/order_spec_test.go` 新增「有规格商品缺规格→400」用例，先 RED 后 GREEN；`go test ./...` 全绿。
**Files:** `backend/api/handler/order.go`、`backend/api/handler/order_spec_test.go`
**Scope:** S

### Task 2 — 【P1】福利金扣减原子化，消除 TOCTOU 双花
**问题：** `order.go:121` 在事务外读 `user.RewardBalance` 计算 `deductAmount`，`:170-176` 用 `reward_balance - ?` 无条件扣减。两个并发订单可基于同一份余额各自扣减，把余额扣成负数。
**做法：** 扣减语句加条件 `Where("id = ? AND reward_balance >= ?", userID, deductAmount)`，检查 `RowsAffected == 0` 则回滚并返回错误（余额不足/被并发改动）。保持 `clause.Locking` 不变。
**Acceptance:**
- [ ] 扣减为「条件更新 + RowsAffected 校验」，余额不足时整单回滚、不创建订单。
- [ ] 正常使用福利金抵扣路径行为不变。
**Verification:** [ ] `order.go` 对应分支断言 `RowsAffected`；新增/补充测试覆盖「余额恰好够」与「条件不满足回滚」；`go test ./...` 全绿。
**Files:** `backend/api/handler/order.go`、相关 `_test.go`
**Scope:** S

### Task 3 — 【P1】下单必须带明细 + order_type 白名单
**问题一：** `order.go:70` 当 `items` 为空时跳过服务端计价，直接信任客户端 `amount`。客户端可空明细传任意金额。
**问题二：** `orderType` 接受任意字符串，仅 `dine_in`/`delivery` 合法。
**做法：** `items` 为空 → 400「订单明细不能为空」；`orderType` 非 `dine_in`/`delivery` → 400。
**Acceptance:**
- [ ] 空 `items` → 400，订单不创建。
- [ ] `order_type` 非法值 → 400。
- [ ] 合法 `dine_in`/`delivery` 带明细 → 行为不变。
**Verification:** [ ] `order_type_test.go` 补「非法 order_type→400」「空 items→400」用例；`go test ./...` 全绿。
**Files:** `backend/api/handler/order.go`、`backend/api/handler/order_type_test.go`
**Scope:** S

### Task 4 — 【P2】GetShop 公开 DTO，不泄露分佣费率
**问题：** 公开的 `GetShop`（小程序 order-confirm 调用）直接返回整个 `models.Shop`，泄露 `reward_rate_self/level1/level2`、`reward_exclude_categories` 等分佣经济结构。
**做法：** `GetShop` 返回精简 DTO：仅保留前端实际需要的展示字段（name/description/address/phone/hours/logo/status）+ 抵扣计算所需的 `reward_ceiling`，剔除分佣费率字段。商户端 `GetShops` 不变（鉴权后可见全量）。
**Acceptance:**
- [ ] `GET /shops/:id` 响应不含 `reward_rate_*`、`reward_exclude_categories`。
- [ ] 仍含 `reward_ceiling`（前端 `refreshCartDisplay` 依赖）及展示字段。
- [ ] 前端 order-confirm 抵扣计算不受影响。
**Verification:** [ ] 新增/补充 handler 测试断言响应字段集合；`go test ./...` 全绿。
**Files:** `backend/api/handler/shop.go`、相关 `_test.go`
**Scope:** S

### Task 5 — 【P2】order-confirm 列表 wx:key 修正
**问题：** `order-confirm/index.wxml:20` 用 `wx:key="id"`，但购物车项无 `id` 字段，只有 `key`（`${productId}_${specId}`）。导致 WeChat 列表 diff 失效告警。
**做法：** 改为 `wx:key="key"`。
**Acceptance:** [ ] order-confirm 明细 `wx:for` 使用 `wx:key="key"`。
**Verification:** [ ] 前端测试断言 wxml 使用 `wx:key="key"`；`jest` 全绿。
**Files:** `frontend/pages/order-confirm/index.wxml`、`frontend/__tests__/order-confirm-type.test.js`
**Scope:** S

### Task 6 — 【P2】允许创建「下架」规格
**问题：** `product.go:220-223` 强制 `status 0→1`，商户无法创建初始即下架的规格。
**做法：** `CreateProductSpecRequest.Status` 改 `*int` 指针；nil → 默认 1，显式 0 → 下架。与 `UpdateProductSpecRequest` 一致。
**Acceptance:**
- [ ] 显式传 `status:0` 创建出下架规格。
- [ ] 不传 status → 仍默认上架（行为不变）。
**Verification:** [ ] `merchant_spec_test.go` 补「创建下架规格」用例；`go test ./...` 全绿。admin 端如有对应入参，`vitest` 全绿。
**Files:** `backend/api/handler/product.go`、`backend/api/handler/merchant_spec_test.go`
**Scope:** S

### Checkpoint（上线前）
- [ ] 三套测试全绿：`go test ./...`、`jest`、`vitest`。
- [ ] 人工核对：有规格商品缺规格下单被拒；福利金并发不出负；GetShop 响应无分佣字段。
- [ ] 重新出 GO/NO-GO 结论。

---

## 不在本次范围（已知、单独决策）
- **零元自动支付路径**（`order.go:193-213`，main 已有、未测）：test-engineer 评为 Critical，但属**既有行为**。建议补**特征化回归测试**（不改行为）——可作为 Checkpoint 后的独立小任务，或单列。本计划默认**不动其逻辑**。
- **邀请奖励 TOCTOU**（`processInviteReward`，main 已有）：异步入账、非下单关键路径，本次不改。
- **seed 弱口令**（`cmd/seed`）：仅本地开发种子，生产不使用，不在上线范围。

## 风险
| 风险 | 缓解 |
|------|------|
| T1 多一次规格存在性查询 | 仅 `spec_id:0` 分支触发，且按 product_id 走索引，开销可忽略 |
| T4 改公开响应结构可能影响前端字段 | 前端仅用 reward_ceiling + 展示字段，DTO 保留这些；测试断言字段集合 |
| 同文件多任务连改 | 按 T1→T4 顺序逐个提交，降低冲突 |

## 回滚
- 分支未 push / 未 merge / 未部署：首要回滚 = 不合并。
- AutoMigrate 仅增列，幂等无害；无需 down 迁移。
