# Implementation Plan: 订单运营台 — Ship Remediation R3

## Overview
Round-3 ship review returned **NO-GO** on two findings, both introduced/exposed by the
R2 work itself:
- **HIGH (security):** the new rate limiter (T7-R2) is bypassable via `X-Forwarded-For`
  spoofing — `main.go` uses bare `gin.Default()` with no `SetTrustedProxies`, so
  `c.ClientIP()` honors a client-supplied header. The brute-force control is cosmetic.
- **MEDIUM (security/financial):** the redispatch `order.Status==2` gate is a read-then-act,
  not part of the atomic claim. A concurrent `2→4` cancel — a path **T5-R2 newly made
  first-class** — during the `CalculatePrice` window lets a billable courier be placed on a
  cancelled order.

This round closes both blockers, hardens the limiter (eviction + observability), and fills
the test gaps the test-engineer flagged (incl. a characterization test for the accepted
limbo `status=0` risk so future changes can't silently reopen the dispatch race).

## Architecture Decisions
- **Trusted proxies via config, default deny-all-proxies.** Add `TrustedProxies []string`
  to `ServerConfig` (env `TRUSTED_PROXIES`, comma-separated, mirroring `AllowedOrigins`).
  `main.go` calls `r.SetTrustedProxies(cfg.Server.TrustedProxies)`. Empty list ⇒ trust no
  proxies (gin then derives the IP from the TCP peer, ignoring spoofed XFF). The exact
  Railway edge CIDR is a deploy-time value (see Open Questions) — the mechanism ships now;
  the value is set in the Railway env.
- **Enforce the status invariant in the DB, not in Go.** Fold `order.Status=2` into the
  redispatch atomic claim via a correlated subquery so the database — not a stale in-memory
  read — rejects a mid-flight-cancelled order. `RowsAffected==0 → 409`.
- **Keep audit-after-commit ordering.** Do NOT move `logOrderAction` into the action's
  transaction — a failed audit insert must not roll back a legitimate 出餐/改状态. Only make
  the failure observable (`log.Printf`).
- **No behavior change to the limbo `status=0` path.** It stays a documented risk; this round
  only adds a characterization test pinning the current 400 response so the claimable set
  `{-1,60}` can't be widened by accident (that is what reopened the race in dropped T3).

## Dependency Graph
```
config.ServerConfig.TrustedProxies ── main.go SetTrustedProxies ── T1 (limiter not spoofable)
ratelimit.go (eviction, now-seam) ──────────────────────────────── T3 (limiter hardening+tests)
merchant_order.go RedispatchOrder claim ───────────────────────── T2 (status in atomic claim)
logOrderAction ────────────────────────────────────────────────── T4 (observable audit failure)
merchant_order_test.go / ratelimit_test.go ────────────────────── T5/T3 (coverage)
```
T1–T4 are independent of each other; T5 depends on T2 (status-claim behavior) being in place.

## Task List

### Phase 1 — GO blockers (security)

- [ ] **T1 ★blocker (HIGH)** Trusted-proxy config so `ClientIP()` can't be spoofed — M
  - **Description:** Add `TrustedProxies []string` to `ServerConfig` + `TRUSTED_PROXIES` env
    parse (comma-split, trim, like `AllowedOrigins`). In `main.go`, after `gin.Default()`,
    call `r.SetTrustedProxies(cfg.Server.TrustedProxies)` and log a warning if the list is
    empty (so the deploy gap is visible). Document the Railway env var.
  - **Acceptance criteria:**
    - [ ] With trusted proxies = empty, a request carrying a spoofed `X-Forwarded-For` does
      **not** change the limiter key (i.e. repeated spoofed-XFF requests from one peer hit the
      same bucket and get 429).
    - [ ] `TRUSTED_PROXIES` env populates `cfg.Server.TrustedProxies` (config test).
  - **Verification:** `cd backend && go test ./middleware/ ./config/`; `go build ./...`;
    manual: grep `main.go` for `SetTrustedProxies`.
  - **Files:** `backend/config/config.go`, `backend/config/config_test.go`, `backend/main.go`,
    `backend/middleware/ratelimit_test.go` (spoof test via an engine with no trusted proxies).
  - **Dependencies:** None.

- [ ] **T2 ★blocker (MED, financial)** Redispatch atomic claim enforces `order.Status=2` — M
  - **Description:** Add a correlated condition to the claim UPDATE so it only matches while
    the parent order is still status 2:
    `Where("id = ? AND shansong_status IN ? AND order_id IN (?)", od.ID, []int{-1,60},
    config.DB.Model(&models.Order{}).Select("id").Where("id = ? AND status = 2", order.ID))`.
    `RowsAffected==0` already returns 409.
  - **Acceptance criteria:**
    - [ ] A redispatch whose order is cancelled (`status→4`) **after** the status read but
      **before** the claim places **no** courier order (0 `orderPlace` calls) and returns
      409/400. Proven with a `quoteDelay` race test that flips `status=4` mid-quote (mirrors
      `ConcurrentDispatchesOnce`).
    - [ ] The existing happy-path redispatch + `ConcurrentDispatchesOnce` still pass.
  - **Verification:** `cd backend && go test ./api/handler/ -run 'Redispatch' -count=5`.
  - **Files:** `backend/api/handler/merchant_order.go`, `backend/api/handler/merchant_order_test.go`.
  - **Dependencies:** None.

### Checkpoint A (GO gate)
- [ ] `go test ./...` + `admin npm test && npm run build` all green.
- [ ] Spoofed XFF cannot reset the rate limit; cancelled-mid-flight order cannot be dispatched.
- [ ] Ship can flip both blockers to resolved.

### Phase 2 — Robustness (recommended code fixes)

- [ ] **T3** Rate-limiter hardening: evict empty keys + cover the sliding window — S
  - **Description:** In `allow()`, after trimming, `if len(kept)==0 { delete(rl.hits,key) }`
    (and still return true/append for a live hit). No janitor needed once T1 stops
    attacker-minted keys.
  - **Acceptance criteria:**
    - [ ] Window-reset test (inject `now`): over-limit → advance `now` past `window` →
      allowed again. (Uses the existing unused `now` seam — zero sleeps.)
    - [ ] After a key's hits all expire and it's checked, it is removed from the map.
    - [ ] `ByUserID` keys two different users independently and falls back to IP when no
      `user_id` is set.
  - **Verification:** `cd backend && go test ./middleware/`.
  - **Files:** `backend/middleware/ratelimit.go`, `backend/middleware/ratelimit_test.go`.
  - **Dependencies:** None (complements T1).

- [ ] **T4** Audit-log write failures are observable — S
  - **Description:** In `logOrderAction`, capture `config.DB.Create(...).Error` and
    `log.Printf` it (keep audit-after-commit, do not fail the action). Add a one-line comment
    stating the deliberate fire-and-forget ordering.
  - **Acceptance criteria:**
    - [ ] A failing audit insert logs an error and the action still returns success
      (verified by inspection; no behavior regression in existing audit tests).
  - **Verification:** `cd backend && go build ./... && go test ./api/handler/ -run 'Audit|Prepare|UpdateOrderStatus'`.
  - **Files:** `backend/api/handler/merchant_order.go`.
  - **Dependencies:** None.

### Phase 3 — Test coverage

- [ ] **T5** Handler coverage gaps — S/M (depends on T2)
  - **Description:** Add the missing handler tests the review flagged.
  - **Acceptance criteria:**
    - [ ] **Limbo characterization:** a delivery order at `shansong_status=0` → redispatch →
      400 "order not in a re-dispatchable state" (pins the accepted risk; guards the
      claimable set `{-1,60}`).
    - [ ] **Redispatch audit:** the success test asserts an `action="redispatch"` row was written.
    - [ ] **Transition whitelist table test:** legal `1→4`, `2→3`, `2→4` succeed; illegal
      `2→1`, `2→2`, `3→1`, `3→4`, `4→2`, `4→3` each return 400 with status unchanged.
  - **Verification:** `cd backend && go test ./api/handler/`.
  - **Files:** `backend/api/handler/merchant_order_test.go`.
  - **Dependencies:** T2.

### Checkpoint Complete
- [ ] `go test ./...` (incl. `-count=5` on Redispatch) + `admin npm test && npm run build` green.
- [ ] Re-run `/ship` → expect GO with only acknowledged risks remaining.
- [ ] End-to-end: spoofed XFF throttled; cancel-mid-redispatch blocked; audit rows written; limbo 400.

## Risks and Mitigations
| Risk | Impact | Mitigation |
|------|--------|------------|
| `TRUSTED_PROXIES` not set in Railway ⇒ `ClientIP` may be the edge IP (all clients share a bucket) | Med | Ship mechanism + warning log; set the env to Railway's edge CIDR at deploy (Open Question). Empty = deny-all-proxies, which is safe-but-coarse, never "trust all". |
| Cancel-race test is timing-based and could flake | Low | Use the proven `quoteDelay` widening + `-count=5`, same pattern as `ConcurrentDispatchesOnce`. |
| Correlated subquery in the claim behaves differently across SQLite/PG | Low | Tests run on the real Postgres test DB (CI `REQUIRE_TEST_DB=1`). |
| Limiter still per-instance | Low | Documented; move to Redis before horizontal scale (not this round). |

## Open Questions
- **Railway edge proxy CIDR / hop count** for `TRUSTED_PROXIES` — operational value, set at
  deploy. Until set, limiter keys on the deny-all-proxies result (TCP peer).
- **Pre-existing `CreateMerchantOrder` creates `Status:2` with a client `amount`** (on `main`,
  out of this branch's scope) — confirm whether merchant-created paid orders are an intended
  trusted POS path; if not, constrain/audit `amount` separately.
- **Limbo `status=0` reconciliation sweep** (time-based re-claim) — deferred; this round only
  characterizes current behavior.

## Constraints (carried from prior rounds — MUST hold)
- Work on branch `feat/merchant-order-board`; **never** touch/stage `frontend/config.js`.
- Git hygiene: stage only each task's files; no `git add -A`; never stage `.claude/`.
- One commit per task, RED→GREEN→regression→build→commit; footer
  `Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>`.
- `-race` unusable locally (cgo off); rely on `-count=N` locally and CI for real `-race`.
