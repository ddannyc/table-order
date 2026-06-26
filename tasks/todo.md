# Todo：/ship 评审整改（双模式下单上线前修复）

详见 `tasks/plan.md`。一任务一提交，TDD（RED→GREEN→回归）。
基线：绝不 git add `frontend/config.js` / `.claude/` / `specs/shansong-delivery.md`。

## P0 阻断项
- [x] **T1** 规格必选守卫：有规格商品传 `spec_id:0` → 400，堵低价下单 — S【blocker】 ✅

## P1 资金路径
- [x] **T2** 福利金扣减原子化：条件更新 + RowsAffected 校验，消除 TOCTOU 双花 — S ✅
- [ ] **T3** 下单必须带明细 + order_type 白名单（空 items / 非法 type → 400）— S

## P2 正确性 / 信息泄露
- [ ] **T4** GetShop 公开 DTO：剔除 reward_rate_* 分佣字段，保留 reward_ceiling — S
- [ ] **T5** order-confirm 列表 `wx:key="id"` → `"key"` — S
- [ ] **T6** 允许创建「下架」规格（Status 改 *int 指针）— S

## Checkpoint
- [ ] `go test ./...` / `jest` / `vitest` 三套全绿
- [ ] 人工核对三项资金/泄露修复；重出 GO/NO-GO

## 不做（单独决策）
- 零元自动支付路径（main 既有、未测）：仅建议补特征化回归测试，不改逻辑
- 邀请奖励 TOCTOU（main 既有，异步非关键路径）
- seed 弱口令（仅本地开发种子）
