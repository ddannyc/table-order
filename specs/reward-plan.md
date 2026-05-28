# Plan: 福利金币分佣系统

## Architecture

```
order.go (handler) ──→ services/reward.go ──→ models (DB)
                      ├─ DistributeReward()    ├─ User
                      ├─ CheckInactivity()     ├─ WalletLog
                      └─ GetExpiredRewards()   └─ RewardLog (NEW)
```

Reward 分发在订单创建后异步执行 (goroutine). 不阻塞支付流程.

## Implementation Order (dependency chain)

```
Step 1: Model layer (DB schema)
  ├─ user.go: +phone_verified, +last_consume_at, +reward_paused_at
  ├─ shop.go: +reward_rate_self, +reward_rate_level1, +reward_rate_level2, +reward_ceiling, +reward_exclude_categories
  ├─ reward_log.go: NEW — tracks each reward issuance with expires_at
  └─ AutoMigrate update

Step 2: Service layer (business logic)
  └─ services/reward.go: DistributeReward, CheckAndPauseInactivity, SweepExpiredRewards

Step 3: Handler layer (API)
  ├─ auth.go: +phone verification → set phone_verified=true
  ├─ order.go: +call DistributeReward after payment, +update last_consume_at
  └─ reward.go: NEW — GET /reward/balance, GET /reward/logs, GET /reward/expiry-info

Step 4: Router
  └─ router.go: +reward routes

Step 5: Frontend — profile page
  └─ profile/index: +reward balance, +reward log tab, +reward paused warning

Step 6: Frontend — order confirm
  └─ order-confirm/index: reward deduction toggle (max 50%)

Step 7: Frontend — invite page
  └─ invite/index: +reward rate display (self 3%, level-1 10%, level-2 4%)
```

## Risk & Mitigation

| Risk | Mitigation |
|------|-----------|
| Reward distribution fails after order committed | Use goroutine + retry, not blocking. Order succeeds regardless |
| Reward expiry sweep misses records | Query filter: `WHERE expires_at < NOW() AND expired = false` |
| 90-day inactivity check slows order creation | Single SELECT on user row (no join), lightweight |
| 50% ceiling bypass via multiple partial orders | Per-order ceiling, not cumulative |

## Verification Checkpoints

1. After Step 1: `go run main.go` — AutoMigrate creates new tables/columns
2. After Step 2: `go test ./services/` — reward calculation tests pass
3. After Step 3: curl API endpoints, verify reward distribution
4. After Step 5-7: WeChat DevTools preview, manual flow test
