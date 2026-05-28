# Spec: 福利金币分佣系统 (Reward Coin System)

## Objective

Build 会员推广分佣系统 — 用户微信授权 + 手机号验证后自动成为"电子股东"，享三级返利（自购3%/直推10%/间推4%），返利以"福利金币"形式发放，仅消费抵扣不可提现。

用户故事:
- 用户扫码进店 → 授权登录 → 自动成为电子股东
- 用户分享专属二维码 → 好友扫码锁定关系 → 好友下单后上级获金币奖励
- 用户下单时可用金币抵扣 50% 金额
- 金币 180 天过期，连续 90 天无消费暂停返利资格

## Tech Stack

- Frontend: 微信原生小程序 (WeChat Mini-Program)
- Backend: Go 1.21 + Gin + GORM + PostgreSQL + Redis
- Auth: JWT (HS256)

## Commands

```
Backend: cd backend && go run main.go
DB:     docker compose up -d (PostgreSQL:5433, Redis:6379)
Test:   cd backend && go test ./...
```

## Project Structure (affected files)

```
backend/
  models/
    user.go              → + fields: phone_verified, reward_paused_at, last_consume_at
    shop.go              → + fields: reward_rate_level1/level2, reward_ceiling
    reward_log.go        → NEW: reward transaction log (type: self/invite_level1/invite_level2)
    reward_expiry.go     → NEW: reward expiry tracking
  api/
    handler/
      auth.go            → + phone verification → auto-activate shareholder
      order.go           → + 3-tier reward distribution after order complete
      reward.go          → NEW: reward balance, logs, expiry info endpoints
    router/router.go     → + reward routes
  services/
    reward.go            → NEW: reward calculation, distribution, expiry sweep logic
  middleware/
    auth.go              → + phone_verified check for reward eligibility
frontend/
  api/index.js           → + reward API calls
  pages/
    profile/index.js     → + reward balance display, reward log tab
    order-confirm/index.js → + reward deduction toggle (max 50%)
    invite/index.js      → + reward rate display, tier info
  app.json               → + "我的推广码" page (share page)
```

## Code Style

Existing Go patterns:
- Handler: thin layer, calls service for logic
- Service: pure business logic, DB access via config.DB
- Model: GORM struct with json tags, TableName()
- Error handling: return early, log.Printf for service errors

Existing JS patterns:
- Page({ data, onLoad, ...methods })
- API calls via require('../../api/index.js')
- wx.getStorageSync for persistent state
- No formal state management

## Testing Strategy

- Go: unit tests for reward calculation logic (services/reward_test.go)
- Go: integration tests for order → reward distribution flow
- Manual: WeChat DevTools frontend testing

## Boundaries

- Always: Validate reward distribution in DB transaction, log all reward changes
- Ask first: Changing reward rates, adding new reward tiers, modifying expiry logic
- Never: Allow reward withdrawal/cash-out, create 3+ tier commission chains, store plaintext phone numbers

## Success Criteria

1. New user auto-activates as "电子股东" on phone verification
2. Self-purchase: 3% reward credited on order completion
3. Level-1 invite: 10% reward to inviter on order completion
4. Level-2 invite: 4% reward to inviter's inviter on order completion
5. Reward balance capped at 50% of order amount for deduction
6. Reward expires 180 days from issue
7. 90-day inactivity pauses reward accrual (not balance)
8. Invite binding is permanent — one scan, lifetime lock
9. Each user has dedicated invite QR code page

## Decisions

1. Phone verification → `wx.getPhoneNumber` (小程序原生能力, 无短信验证码)
2. Reward trigger → `status=2` (paid). 系统无确认收货流程, 支付即完成. MVP 简化
3. Inactivity check → 实时. 下单时检查 `last_consume_at`, 不引入 cron job
4. Reward expiry → 统一过期. 每条 reward_log 记录 `expires_at`, 查询时过滤已过期
5. Exclusion → 店铺级配置. shop 表加 `reward_exclude_categories` (JSON array of category names)
