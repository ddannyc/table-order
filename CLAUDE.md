# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Restaurant member referral mini-program ("餐饮店扫码会员裂变小程序"). Go/Gin backend + uni-app Vue3 frontend. Users scan QR codes at restaurants, become members, earn 10% cashback on purchases, and 5% referral rewards when inviting friends.

## Architecture

```
table-order/
├── backend/          # Go/Gin API server
│   ├── api/
│   │   ├── handler/  # HTTP handlers (auth, order, wallet, invite, shop, qrcode, admin, merchant)
│   │   └── router/   # Route definitions
│   ├── config/       # DB/Redis init, config.yaml loading, AppConfig global
│   ├── middleware/   # Auth, JWT validation
│   ├── models/       # GORM models (User, Shop, Order, WalletLog, etc.)
│   ├── services/     # Business logic (WeChat API)
│   └── utils/        # Shared utilities (ParseUint)
├── frontend/         # uni-app Vue3 mini-program
│   ├── src/
│   │   ├── api/      # API client functions
│   │   ├── pages/    # Page components (login, home, wallet, invite, orders, scan)
│   │   └── static/    # Tab bar icons
│   └── dist/build/mp-weixin/  # Built WeChat mini-program output
├── docker-compose.yml
└── 餐饮店扫码会员裂变小程序_prd.md
```

## Common Commands

### Backend
```bash
cd backend
go run main.go                        # Start server on :8080
go build ./...                        # Build
go test ./...                         # Run tests
```

### Frontend
```bash
cd frontend
npm run dev:h5                        # H5 dev server (localhost:5174)
npm run dev:mp-weixin                 # WeChat mini-program dev
npm run build:mp-weixin              # Build WeChat mini-program for dev
npm run build:mp-weixin:prod         # Build WeChat mini-program for prod
```

### Database
```bash
# PostgreSQL (local)
psql -U postgres -d table_order -h localhost -p 5432

# Docker postgres
docker exec -it <container> psql -U postgres -d table_order
```

## Key Technical Details

### API Base URL
- Dev: `http://localhost:8080/api` (for H5)
- WeChat mini-program: Use machine IP or domain — `localhost` doesn't work from mobile device/emulator
- Configured via `vite.config.js` define plugin, not runtime env

### Database
- PostgreSQL on port 5432 (local) or 5433 (docker)
- GORM with `github.com/example/table-order` module path
- Column names use snake_case (e.g., `open_id`, not `openid`)
- Numeric decimals use `numeric(12,2)`

### Auth
- JWT token with `user_id`, `openid`, `role` claims
- Token signed with `config.AppConfig.JWT.Secret`
- Roles: 0=user, 1=merchant, 2=admin
- WeChat auth: real API when `appid`/`appsecret` configured, fallback to mock `mock_openid_*` for dev

### Order Flow
1. Create order → atomic transaction (order + reward balance + wallet log)
2. After commit → async `processInviteReward` for inviter
3. Reward: 10% of amount (shop.RewardRate)
4. Invite reward: 5% of amount (shop.InviteRate)

### Pagination
- `GetOrders` and `GetWalletLogs` use `page` and `page_size` query params
- Response format: `{orders/logs: [], total, page, page_size}`

## Build Output

- WeChat mini-program: `frontend/dist/build/mp-weixin/`
- Copy to Windows local path for WeChat dev tools (WSL path access issue)
- Tab bar icons must be 81x81 PNG in `src/static/` before build