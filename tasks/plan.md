# 实施计划：闪送外卖配送对接

> 基于 `specs/shansong-delivery.md`。**双模式合并已落地大量脚手架**，本计划只做真正的闪送集成。

## 概述
`feat/dual-mode-ordering`（已并入 main）已完成「堂食/外卖」下单方式拆分。本计划在其之上对接**闪送开放平台**：配送详情落库、实时运费报价、支付成功后自动派单、状态回调入库，以及前端取地址+取坐标+显示运费+展示配送状态的完整链路。

所有修复遵循项目 TDD 约定（RED → GREEN → 回归 → 一任务一提交）。后端 `cd backend && go test ./...`（需本地 Postgres），前端 `node node_modules/jest/bin/jest.js`，admin `node node_modules/vitest/vitest.mjs run`。

**基线纪律：** 绝不 `git add` `frontend/config.js`、`.claude/`、`specs/shansong-delivery.md`；每个提交只暂存该任务相关文件。`config.yaml.example` 只放占位密钥；真实闪送凭据仅经环境变量 / 被 gitignore 的 `config.yaml`，绝不入库。

## 已就绪（双模式合并，无需重做）
- `Order.OrderType`（dine_in|delivery）字段 + `CreateOrderRequest.OrderType` 白名单 + delivery 时 `table_no` 可空。
- `Shop.Latitude/Longitude` 字段 + `UpdateShop` 接受经纬度（**寄件方坐标已可由商家维护** → spec Open Q1 数据层已解）。
- `publicShopDTO`（含 lat/lng，剔除分佣费率）+ `GetShop` / `ResolveDeliveryShop` + `/api/delivery/shop` 路由。
- `app.json` `requiredPrivateInfos:["chooseAddress"]`；前端 `resolveDeliveryShop()`、`getLastDeliveryAddress/setLastDeliveryAddress`、order-confirm 读取 orderType + 展示地址卡片。

## 本期已确认的关键决策（来自人工确认）
1. **目的地坐标 = 前端 `wx.getLocation`（gcj02）取用户当前坐标**，与 `wx.chooseAddress` 的文本地址组合提交。**不引入后端地理编码依赖。**
2. **取消/退款本期不做**：闪送 service 可留 `CancelOrder` 方法，但不接路由、不接 UI。
3. **配送状态展示纳入本期**：回调写库 + `GetOrders/GetOrder` 回传 delivery 信息 + 我的订单展示状态文案。
4. **闪送测试凭据已就绪，本期含真机联调** checkpoint。

## 分级
- **P1 后端基础 + 报价**：模型 / 配置 / service / 报价接口 —— 履约链路前置。
- **P2 下单与履约**：下单落库含运费 / 支付后派单 / 状态回调。
- **P3 前端 + 状态展示**：取地址取坐标报价下单 / 我的订单展示。
- **P4 联调**：测试环境真机全链路。

## 依赖关系 / 执行序
T1（基础设施）→ T2（service）→ T3（报价接口）→ T4（下单落库）→ T5（派单钩子）→ T6（回调）→ T7（前端下单）→ T8（状态展示）→ 联调 checkpoint。
- T7 依赖 T3（报价）+ T4（下单 payload）。
- T8 依赖 T4（OrderDelivery）+ T6（回调写入的 status）。

---

## 任务列表

### Task 1 —【P1】配送基础设施：OrderDelivery 模型 + 迁移 + ShansongConfig
**做法：** 新增 `models/order_delivery.go`（按 spec Data Model：1:1 `OrderID` 唯一索引、收件人/地址、`DeliveryFee`、`ShansongOrderNo`、`ShansongStatus`，**补 `RecipientLat/RecipientLng`** 供报价/派单）；`database.go` 注册迁移；`config.go` 新增 `ShansongConfig`（clientid/appSecret/baseURL/notifyURL）+ 环境变量读取；`config.yaml.example` 补占位样例。
**Acceptance:**
- [ ] `AutoMigrate` 包含 `OrderDelivery`，`order_id` 唯一索引。
- [ ] `ShansongConfig` 经环境变量可覆盖；`config.yaml.example` 仅占位、无真实密钥。
**Verification:** [ ] 迁移冒烟测试建表成功；`go build ./...` 通过。
**Files:** `backend/models/order_delivery.go`、`backend/config/database.go`、`backend/config/config.go`、`backend/config/config.yaml.example`
**Scope:** S

### Task 2 —【P1】闪送 service 客户端：签名 + 报价 + 下单
**做法：** 新增 `services/shansong.go`：`ShansongClient`（ClientID/AppSecret/BaseURL）；`CalculatePrice(QuoteRequest)`、`CreateOrder(...)`、`CancelOrder(...)`（留方法本期不接路由）；MD5 签名按官方文档；闪送状态码 → 中文文案映射。HTTP 经可注入的 client 以便 mock。
**Acceptance:**
- [ ] 已知输入 → 期望签名（确定性单测）。
- [ ] 报价/下单请求体构造正确、响应解析正确（mock HTTP）。
- [ ] 网络/业务错误返回 error，不 panic。
**Verification:** [ ] `services/shansong_test.go` 覆盖签名 + 报价 + 下单（mock HTTP）；`go test ./...` 全绿。
**Files:** `backend/services/shansong.go`、`backend/services/shansong_test.go`
**Scope:** M

### Task 3 —【P1】报价接口 `POST /api/delivery/quote`
**做法：** 新增 `handler/delivery.go`：入参 `shop_id` + 收件地址 + 收件坐标（lat/lng）；取 `Shop` 寄件坐标（为 0 → 400「门店未配置坐标」）；调 `CalculatePrice`；返回 `delivery_fee` + **后端签名的报价凭证**（HMAC，内含 fee+坐标+过期时间），下单时回传校验，杜绝客户端伪造运费。报价失败/超距 → 明确错误，前端据此禁止下单。注册鉴权路由。
**Acceptance:**
- [ ] 合法入参 → 返回 `delivery_fee` + 报价凭证。
- [ ] 门店坐标缺失 / 报价失败 → 4xx + 明确 message。
- [ ] 报价凭证可被后端验签且带过期时间。
**Verification:** [ ] `handler/delivery_test.go`：入参校验、运费透传、坐标缺失、凭证签发（mock service）；`go test ./...` 全绿。
**Files:** `backend/api/handler/delivery.go`、`backend/api/router/router.go`、`backend/api/handler/delivery_test.go`
**Scope:** M

### Task 4 —【P2】CreateOrder delivery 分支落库 + 应付含运费
**做法：** `CreateOrderRequest` 新增 `Delivery`（收件人/电话/省市区详细/lat/lng）+ `QuoteToken`（报价凭证）。delivery 单：**校验报价凭证**（验签 + 未过期）取信任的 `delivery_fee`，**不信任客户端任意运费**；落 `OrderDelivery`（同事务）。金额口径：`order.Amount` 仍 = 菜品额 − 福利金抵扣（**返利/福利金基数不变**）；**实付 = order.Amount + delivery_fee**（prepay 金额与零元判断都用实付，避免有运费却走零元自动支付）。堂食分支完全不变。
**Acceptance:**
- [ ] delivery 单：`OrderDelivery` 落库（地址+坐标+运费），`table_no` 可空。
- [ ] 实付 = 菜品净额 + 运费；prepay 金额含运费；菜品被全额抵扣但有运费时**不**走零元路径。
- [ ] 报价凭证无效/过期 → 400，订单不创建。
- [ ] 返利按 `order.Amount`（不含运费）；堂食回归不变。
**Verification:** [ ] `handler/order_delivery_test.go`：落库、含运费实付、凭证校验、运费不进返利、堂食回归；`go test ./...` 全绿。
**Files:** `backend/api/handler/order.go`、`backend/api/handler/order_delivery_test.go`
**Scope:** M

### Task 5 —【P2】支付成功后自动派单
**做法：** 新增 `services.DispatchShansong(orderID)`：读 `OrderDelivery` → 调 `CreateOrder` → 落 `ShansongOrderNo`/初始 `ShansongStatus`；失败仅记日志不阻塞。在 `wechatpay_notify.go` 支付成功分支、以及 `order.go` 零元自动支付分支（delivery 单理论不会零元，但稳妥起见判 OrderType）后 `go services.DispatchShansong(order.ID)`。仅 delivery 单派单。
**Acceptance:**
- [ ] 支付成功的 delivery 单触发派单，`ShansongOrderNo` 落库。
- [ ] 派单失败只记日志，不影响支付回调返回 / 不回滚订单。
- [ ] 堂食单不派单。
**Verification:** [ ] `services/shansong_dispatch_test.go`（mock 闪送）+ 回调钩子断言；`go test ./...` 全绿。
**Files:** `backend/services/shansong.go`（或新 `shansong_dispatch.go`）、`backend/api/handler/wechatpay_notify.go`、`backend/api/handler/order.go`、对应 `_test.go`
**Scope:** M

### Task 6 —【P2】闪送状态回调 `POST /api/shansong/callback`
**做法：** 新增 `handler/shansong_notify.go`：验签（失败 → 拒绝）；按 `ShansongOrderNo` 更新 `OrderDelivery.ShansongStatus`；**重复回调幂等**；成功**必须返回 `{"status":200}`**（否则闪送重试）。注册公开路由（闪送签名鉴权，非 JWT）。
**Acceptance:**
- [ ] 验签通过 → 更新状态、返回 `{"status":200}`。
- [ ] 验签失败 → 不更新、非 200 业务码。
- [ ] 重复回调幂等（状态不回退、不重复副作用）。
**Verification:** [ ] `handler/shansong_notify_test.go`：验签通过/失败、状态更新、幂等、返回体；`go test ./...` 全绿。
**Files:** `backend/api/handler/shansong_notify.go`、`backend/api/router/router.go`、`backend/api/handler/shansong_notify_test.go`
**Scope:** M

### Task 7 —【P3】前端外卖下单：取地址 + 取坐标 + 报价 + 含运费
**做法：** `app.json` `requiredPrivateInfos` 增 `getLocation`。order-confirm：外卖时「选择收货地址」→ `wx.chooseAddress` 取文本地址 + `wx.getLocation`（gcj02）取坐标 → 缓存默认（复用 `setLastDeliveryAddress`）；调 `getDeliveryQuote` 显示配送费行；应付 = 商品 − 抵扣 + 运费；`createOrder` payload 扩展 `delivery`（地址+坐标）+ `quote_token`。`api/index.js` 新增 `getDeliveryQuote()` 并扩展 `createOrder`。报价失败 → 禁止支付 + 提示。
**Acceptance:**
- [ ] 外卖单 payload 含 `delivery`（含 lat/lng）+ `quote_token`。
- [ ] 应付正确含运费；报价失败禁止下单。
- [ ] 堂食流程不受影响。
**Verification:** [ ] jest：payload 结构、含运费应付、报价失败禁付、堂食回归；`jest` 全绿。
**Files:** `frontend/app.json`、`frontend/pages/order-confirm/index.{js,wxml,wxss}`、`frontend/api/index.js`、`frontend/__tests__/*.test.js`
**Scope:** M

### Task 8 —【P3】我的订单展示配送状态
**做法：** 后端 `GetOrders/GetOrder` 对 delivery 单一并返回 `OrderDelivery`（运费、状态、状态文案、收件信息）。前端 profile/订单列表（及详情如有）展示「外卖 · 配送状态」文案。
**Acceptance:**
- [ ] delivery 单的订单响应含 delivery 信息 + 状态文案。
- [ ] 前端订单项展示配送状态；堂食单不展示配送区块。
**Verification:** [ ] 后端 handler 测试断言响应含 delivery；前端测试断言渲染状态；`go test ./...` + `jest` 全绿。
**Files:** `backend/api/handler/order.go`、`frontend/pages/profile/index.{js,wxml}`（或订单列表页）、对应 `_test.go` / `*.test.js`
**Scope:** M

### Checkpoint（联调，P4）
- [ ] 三套测试全绿：`go test ./...`、`jest`、`vitest`（admin 若无改动则确认不受影响）。
- [ ] 真机/体验版：`chooseAddress` + `getLocation` 授权回填（**需 mp 后台开通 getLocation 接口权限 + 隐私指引**）。
- [ ] 测试环境全链路：选地址 → 报价展示 → 支付 → 闪送派单（`ShansongOrderNo` 落库）→ 回调更新状态 → 我的订单可见。
- [ ] 堂食全链路回归通过；返利/福利金按菜品额不受运费影响。
- [ ] 仓库无任何闪送真实密钥（`git log -p` 抽查 + `config.yaml.example` 仅占位）。
- [ ] 重新出 GO/NO-GO。

---

## 不在本期范围
- **取消 / 退款**（用户或商家取消 → 闪送取消 + 退运费）：明确留作后续。
- **后端用户地址簿**：仅本地缓存默认地址（spec A4）。
- **多门店就近选店**：`ResolveDeliveryShop` 维持单店 stub。

## 风险
| 风险 | 缓解 |
|------|------|
| 报价凭证过期 / 派单时闪送重算运费差额 | 后端签名凭证带过期时间；差额本期由商家承担、不阻塞支付（记日志，列后续优化） |
| `wx.getLocation` 当前坐标 ≠ 收货地址坐标（用户不在收货地） | 即时配送通常在收货地下单；本期接受该近似，超距由报价兜底拒单 |
| getLocation 需 mp 后台开通权限 + 隐私指引 | 联调前在小程序后台配置，纳入 checkpoint |
| 闪送签名 / endpoint 细节以官方文档为准 | service 以可注入 client + mock 单测锁定契约；联调阶段对真实环境校准 |
| 运费被客户端伪造 | `/quote` 返回后端签名凭证，CreateOrder 验签取信任运费，不信任客户端字段 |

## 回滚
- 分支未 push / 未 merge / 未部署：首要回滚 = 不合并。
- `AutoMigrate` 仅增表/增列，幂等无害；无需 down 迁移。
- 闪送派单为支付后异步、失败不阻塞，回滚下单逻辑不影响已支付订单的资金。
