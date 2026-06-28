# Spec: 闪送外卖配送功能

## Objective

为现有「扫码堂食点餐」小程序新增**外卖配送**下单方式，并对接**闪送开放平台 API**完成实时运费报价、自动派单与配送状态回调。下单时用户可在**堂食 / 外卖配送**之间选择；选择外卖时，通过微信 `wx.chooseAddress` 拉取用户在微信中保存的历史收货地址自动填充，并在本地缓存最近一次地址作为默认值。

**用户**: 餐厅顾客 / 会员（外卖场景：堂食桌号不可用，改为收货地址配送）
**核心流程（外卖）**:
浏览菜品 → 加购 → 订单确认页选「外卖配送」→ `wx.chooseAddress` 选地址（或用缓存默认）→ 调闪送实时报价显示配送费 → 微信支付 → 支付成功后后端自动向闪送下单 → 闪送回调更新配送状态

**成功标准**:
1. 堂食流程完全不受影响（回归通过）。
2. 外卖下单可走通：选地址 → 报价 → 支付 → 闪送派单 → 状态回调落库。
3. 闪送凭据、地址等敏感信息不进入版本库。

---

## ASSUMPTIONS（请确认或纠正）

1. 闪送对接采用其**开放平台 OpenAPI**（测试 `open.s.bingex.com` / 生产 `open.ishansong.com`），凭据为 `clientid` + `appSecret`，由商家在闪送应用中心申请，运维通过环境变量/配置注入。
2. 配送费 = **闪送实时报价**全额计入用户应付金额（不做商家补贴/满减，本期不涉及）。
3. 堂食与外卖**并存**，下单时可选；订单新增 `order_type` 字段区分（`dine_in` / `delivery`）。
4. 地址**仅本地缓存最近一次**（`wx.setStorageSync`）做默认填充，**不建后端用户地址簿**。
5. 闪送派单时机 = **微信支付成功后**（`WechatPayNotify` 内，订单 `status→2` 之后异步派单），避免未支付即产生真实运力成本。
6. 闪送报价需要寄件方（门店）与收件方坐标/地址；门店地址与坐标来自 `Shop` 配置（**见 Open Questions Q1**）。
7. `wx.chooseAddress` 仅返回文本地址（省/市/区/详细/姓名/电话/邮编），**不含经纬度**；目的地坐标的获取方式见 Open Questions Q2。
8. 仅小程序端使用，无需 H5/web 端外卖入口。

---

## Tech Stack

| 层次 | 技术 | 说明 |
|------|------|------|
| 前端 | 微信原生小程序 | 复用现有 `frontend/`（weui-wxss + Vant），新增地址选择与下单方式切换 |
| 微信地址 | `wx.chooseAddress` | 需 `app.json` 声明 `requiredPrivateInfos` + 用户隐私保护指引 |
| 后端 | Go (Gin + GORM + PostgreSQL) | 复用现有 `backend/`，新增闪送 service / 模型 / 路由 |
| 闪送 | 闪送开放平台 OpenAPI | 运费预估 / 下单 / 取消 / 状态回调；MD5 签名（以官方文档为准） |
| 支付 | 微信支付 V3（已接入） | 复用 `WechatPayNotify`，在支付成功分支触发闪送派单 |

---

## Project Structure（新增/改动）

```
backend/
├── config/
│   ├── config.go            # 改：新增 ShansongConfig（clientid/appSecret/baseURL/notifyURL）
│   ├── config.yaml.example  # 改：补充 shansong 配置样例（不含真实密钥）
│   └── database.go          # 改：MigrateDB 注册 &models.OrderDelivery{}
├── models/
│   ├── order.go             # 改：Order 新增 OrderType 字段
│   └── order_delivery.go    # 新：OrderDelivery（1:1 配送详情 + 闪送单号/状态）
├── services/
│   ├── shansong.go          # 新：闪送客户端（CalculatePrice/CreateOrder/CancelOrder/签名）
│   └── shansong_test.go     # 新：签名 + 报价/下单请求构造的单测（mock HTTP）
└── api/
    ├── handler/
    │   ├── order.go            # 改：CreateOrderRequest 支持 order_type + 配送信息；TableNo 放开
    │   ├── delivery.go         # 新：POST /api/delivery/quote（实时报价）
    │   ├── shansong_notify.go  # 新：POST /api/shansong/callback（状态回调，返回 status:200）
    │   └── wechatpay_notify.go # 改：支付成功后 delivery 单 → go services.DispatchShansong(orderID)
    └── router/                 # 改：注册 /api/delivery/quote 与 /api/shansong/callback

frontend/
├── app.json                 # 改：requiredPrivateInfos:["chooseAddress"]；隐私指引（小程序后台配置）
├── api/index.js             # 改：getDeliveryQuote()；createOrder() 扩展 payload
└── pages/order-confirm/
    ├── index.js             # 改：堂食/外卖切换、chooseAddress、报价、应付含配送费
    ├── index.wxml           # 改：下单方式选择 + 地址卡片 + 配送费行
    └── index.wxss           # 改：地址卡片样式

specs/shansong-delivery.md   # 本文件
```

---

## Commands

```bash
# 后端
cd backend && go build ./...          # 编译
cd backend && go test ./...           # 全部单测（含 shansong_test.go）
cd backend && go run main.go          # 本地启动（读 config.yaml / 环境变量）

# 前端：微信开发者工具导入 frontend/，编译预览（chooseAddress 需真机/体验版验证）
```

---

## Data Model

```go
// models/order.go —— 新增字段（其余不变）
type Order struct {
    // ...既有字段...
    OrderType string `gorm:"size:16;default:dine_in" json:"order_type"` // dine_in | delivery
}

// models/order_delivery.go —— 新增
type OrderDelivery struct {
    ID             uint      `gorm:"primaryKey" json:"id"`
    OrderID        uint      `gorm:"uniqueIndex" json:"order_id"`
    RecipientName  string    `gorm:"size:64" json:"recipient_name"`
    RecipientPhone string    `gorm:"size:32" json:"recipient_phone"`
    Province       string    `gorm:"size:32" json:"province"`
    City           string    `gorm:"size:32" json:"city"`
    County         string    `gorm:"size:32" json:"county"`
    DetailAddress  string    `gorm:"size:255" json:"detail_address"`
    DeliveryFee    float64   `gorm:"type:numeric(12,2)" json:"delivery_fee"`
    ShansongOrderNo string   `gorm:"size:64;index" json:"shansong_order_no"` // 闪送返回的运单号
    ShansongStatus int       `gorm:"default:0" json:"shansong_status"`        // 闪送配送状态码
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

---

## API 契约

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/delivery/quote` | 用户 JWT | 入参：`shop_id` + 收件地址；出参：`delivery_fee` + 闪送报价凭证（下单时回传） |
| POST | `/api/orders` (改) | 用户 JWT | 新增 `order_type`、`delivery`（地址）、`delivery_fee`、报价凭证；`delivery` 时 `table_no` 可空 |
| POST | `/api/shansong/callback` | 闪送签名 | 闪送状态回调；验签 → 更新 `OrderDelivery.ShansongStatus`；**成功须返回 `{"status":200}`** |

`createOrder` 金额规则：`delivery` 单应付 = 菜品总额（含福利金抵扣）+ `delivery_fee`；运费不参与福利金抵扣、不参与返利计算（返利仍按菜品额 `order.Amount`）。

---

## Code Style

沿用现有约定（见根 `SPEC.md`）。后端示例：

```go
// services/shansong.go
type ShansongClient struct {
    ClientID  string
    AppSecret string
    BaseURL   string // open.s.bingex.com（测试）/ open.ishansong.com（生产）
}

// CalculatePrice 调用闪送运费预估，返回运费与报价凭证（下单时回传）。
func (c *ShansongClient) CalculatePrice(req QuoteRequest) (*QuoteResult, error) { /* 签名 + POST */ }
```

```go
// 闪送派单：仿照 DistributeReward，在支付成功后异步触发，失败仅记日志不阻塞支付回调
go services.DispatchShansong(order.ID)
```

前端：`wx.chooseAddress` 封装在 `pages/order-confirm/index.js`，选完写入 `wx.setStorageSync('last_delivery_address', addr)` 作下次默认；`wx.request` 调用统一走 `api/index.js`，函数签名对外稳定。

---

## Testing Strategy

- **后端单测（Go testing）**：
  1. 闪送签名生成正确（已知输入 → 期望签名）。
  2. `/api/delivery/quote` 入参校验与运费透传（mock 闪送 HTTP）。
  3. `CreateOrder` 的 `delivery` 分支：应付含运费、`table_no` 可空、`OrderDelivery` 落库；堂食分支回归不变。
  4. `/api/shansong/callback` 验签通过/失败、状态更新、返回 `status:200`、重复回调幂等。
- **回归**：现有 `handler_test.go` 全绿（堂食下单、返利、预支付字段不受影响）。
- **手动（真机/体验版）**：`wx.chooseAddress` 授权与回填、报价展示、支付、闪送测试环境派单与回调全链路。

---

## Boundaries

### Always
- 闪送凭据、回调密钥仅经环境变量/`config.yaml`（已被 gitignore）注入；`config.yaml.example` 只放占位。
- 闪送派单仅在**支付成功后**触发；派单失败只记日志、不影响支付回调返回。
- `/api/shansong/callback` 必须验签，且成功返回 `{"status":200}`（否则闪送重试）。
- 堂食代码路径保持原样，外卖为增量分支。

### Ask First
- 给 `Shop` 增加门店坐标/地址字段（报价寄件方所需，见 Q1）。
- 引入新的后端依赖（如地理编码 SDK）。
- 改动 `app.json` 之外的全局配置或 tabBar。
- 运费是否参与返利/福利金口径的任何调整。

### Never
- 不建后端用户地址簿（本期明确仅本地缓存）。
- 不在前端硬编码闪送密钥或调用闪送签名接口（签名只在后端）。
- 不删除/绕过现有支付成功的返利逻辑。
- 不在未支付状态向闪送真实下单。

---

## Success Criteria

1. [ ] `go build ./...` 与 `go test ./...` 通过，新增单测覆盖签名/报价/下单/回调。
2. [ ] 订单确认页可切换堂食/外卖；外卖时 `wx.chooseAddress` 回填地址并缓存默认。
3. [ ] `/api/delivery/quote` 返回闪送实时运费并在页面正确展示、计入应付。
4. [ ] 外卖单支付成功后自动向闪送（测试环境）下单，`OrderDelivery.ShansongOrderNo` 落库。
5. [ ] `/api/shansong/callback` 验签并更新配送状态，返回 `status:200`。
6. [ ] 堂食全链路回归通过，返利/福利金按菜品额计算不受运费影响。
7. [ ] 仓库无任何闪送真实密钥。

---

## Open Questions

1. **门店寄件地址/坐标来源**：闪送报价需寄件方地址与经纬度。当前 `Shop` 模型是否已有地址/坐标？若无，需新增字段并在商家后台维护——是否本期一并做？
2. **目的地经纬度**：`wx.chooseAddress` 不返回经纬度。闪送报价/下单若强制要求收件方坐标，是否需引入逆地理编码（如腾讯位置服务 key），还是闪送 API 接受纯文本地址？需查官方文档确认。
3. **闪送签名/接口细节**：具体下单、预估、取消的 endpoint 路径与签名算法以闪送最新官方文档为准——是否已有商家账号可拿到测试 `clientid`/`appSecret` 用于联调？
4. **配送范围/超距处理**：超出闪送可达范围或报价失败时，订单确认页如何提示（禁止下单 / 仅堂食）？
5. **取消/退款**：用户或商家取消外卖单时，是否需调闪送取消接口并退运费？本期是否纳入？
6. **配送状态展示**：是否需在「我的-订单」展示闪送配送状态/进度？本期范围确认。
```
