# Spec: 商家管理后台 (Merchant Admin Web)

> 状态：草案，待确认。本文件**不替代**根目录 `SPEC.md`（那份是小程序重构）。

## 1. Objective

为餐饮店**商家**提供一个 Web 管理后台，登录后管理自己的店铺、菜品、查看经营数据、配置返利规则、生成桌台二维码。后端 Go API（Gin，`/api/merchant/*`）已大部分存在，本项目主要是**新建 Web 前端** + **补齐少量后端缺口**。

- **用户**：餐厅商家（一个 merchant 账号可拥有多家 shop）。**不含**平台超管（`/api/admin/*`，留作第二阶段）。
- **登录方式**：手机号 + 密码（已有 `POST /api/merchant/login`），返回 JWT（role=1）。
- **成功标准**：商家能登录 → 选择店铺 → 增删改查菜品、上下架 → 编辑店铺信息与返利比例 → 生成/查看桌码 → 在看板看到订单量/营业额。所有功能跑通本地联调（前端 `localhost:5173` ↔ 后端 `localhost:8080`）。

### 范围内 (v1)
- 登录鉴权（必备）
- 菜品管理（增删改查、分类、上下架、图片）
- 数据看板（累计 + 单日统计卡片）
- 店铺信息编辑 + 返利配置
- 桌台二维码生成与查看
- **订单运营台**（待办队列：堂食已支付未出餐 →【出餐】，外卖派单失败/卡太久 →【重新派单】/【改状态】）。设计依据见 `docs/ideas/order-action-board.md`。

### 范围外 (明确不做)
- 平台超管后台（`/api/admin/*`）
- 会员/返利明细报表、提现
- 订单退款/refund（涉及资金流动，单列范围）、闪送失败自动重试自愈、派单失败推送/短信告警（见 idea 文档 Not Doing）
- 商家自助注册引导（注册接口存在，但 UI 仅做登录；如需可后补）

---

## 2. Tech Stack

| 层 | 选型 | 说明 |
|----|------|------|
| 框架 | Vue 3 (`<script setup>`) | Composition API |
| UI 库 | Element Plus | 中后台主流，表格/表单/弹窗开箱即用 |
| 构建 | Vite | 默认 dev server `:5173` |
| 路由 | Vue Router 4 | history 模式，登录守卫 |
| 状态 | Pinia | 仅存 auth（token + merchant + 当前 shopId） |
| HTTP | axios | 封装 baseURL + 拦截器（注入 token / 401 跳登录） |
| 图表 | 无（v1 看板只做统计卡片） | 不引 ECharts |
| 语言 | **JavaScript**（非 TS） | 与仓库现有 JS 技术栈一致（小程序全 JS、无 TS）。如需 TS 请提出 |
| 部署 | 独立 SPA，单独部署（如 Railway/Vercel/Nginx） | 经 CORS 调后端，后端已设 `Access-Control-Allow-Origin: *` |

---

## 3. Project Structure

新建**顶层** `admin/` 目录（独立部署，不放进 `backend/frontend/` 那个空壳）：

```
admin/
├── index.html
├── package.json
├── vite.config.js
├── .env.development        # VITE_API_BASE=http://localhost:8080/api
├── .env.production         # VITE_API_BASE=<railway url>/api
└── src/
    ├── main.js             # 挂载 app + Element Plus + router + pinia
    ├── App.vue
    ├── api/
    │   ├── client.js       # axios 实例 + 拦截器
    │   ├── auth.js         # login
    │   ├── shop.js         # shops CRUD + 返利配置
    │   ├── product.js      # 菜品 CRUD + 上下架
    │   ├── stats.js        # dashboard + stats
    │   ├── qrcode.js       # 桌码 生成/列表
    │   └── order.js        # 订单列表 + 出餐/重新派单/改状态
    ├── stores/
    │   └── auth.js         # token, merchant, currentShopId；localStorage 持久化
    ├── router/
    │   └── index.js        # 路由表 + beforeEach 登录守卫
    ├── layouts/
    │   └── AdminLayout.vue # 侧边栏 + 顶栏 + 店铺切换器 + <router-view>
    ├── views/
    │   ├── Login.vue
    │   ├── Dashboard.vue   # 看板
    │   ├── Products.vue    # 菜品管理（表格 + 编辑弹窗）
    │   ├── ShopSettings.vue# 店铺信息 + 返利配置
    │   ├── QRCodes.vue     # 桌码
    │   └── Orders.vue      # 订单运营台（默认"待处理"队列 + tab + 筛选 + 行内动作）
    └── components/         # 复用组件（按需）
```

---

## 4. API 契约（现有后端，前端据此对接）

所有 `/api/merchant/*` 需请求头 `Authorization: Bearer <token>`。

| 模块 | 方法 & 路径 | 请求 | 响应（关键字段） |
|------|------------|------|-----------------|
| 登录 | `POST /api/merchant/login` | `{phone, password}` | `{token, merchant:{id,phone,name,company,status}}` |
| 店铺列表 | `GET /api/merchant/shops` | — | `Shop[]` |
| 建店铺 | `POST /api/merchant/shops` | `{name,description,address,phone,hours}` | `Shop` |
| 改店铺 | `PUT /api/merchant/shops/:id` | 见 §5 缺口① | `{message}` |
| 菜品列表 | `GET /api/merchant/products` | — | `Product[]`（含全部状态，跨该商家所有店） |
| 建菜品 | `POST /api/merchant/products` | `{shop_id,name,price,description,image,category,status}` | `Product` |
| 改菜品 | `PUT /api/merchant/products/:id` | 见 §5 缺口② | `{message}` |
| 删菜品 | `DELETE /api/merchant/products/:id` | — | `{message}` |
| 看板 | `GET /api/merchant/dashboard` | — | `{shops, total_users, total_orders, total_revenue}` |
| 单日统计 | `GET /api/merchant/stats?date=YYYY-MM-DD&shop_id=` | query | `{new_users, orders, revenue, rewarded}` |
| 桌码列表 | `GET /api/shops/:id/qrcodes` | — | `TableQRCode[]`（**公开**，见缺口③） |
| 生成桌码 | `POST /api/shops/:id/qrcodes` | `{table_no}` | `{...qrcode 图片}`（**公开**，见缺口③） |
| 订单列表 | `GET /api/merchant/orders?shop_id=&date=&status=&type=` | query | `{orders:[Order+delivery], total, revenue, rewarded}`（见缺口⑥，含 delivery 明细 + 分页） |
| 出餐 | `POST /api/merchant/orders/:id/prepare` | — | `{message}`（置 `PreparedAt`，见缺口⑥） |
| 重新派单 | `POST /api/merchant/orders/:id/redispatch` | — | `{message}`（调 `DispatchShansong`，见缺口⑥） |
| 改状态 | `PUT /api/merchant/orders/:id/status` | `{status}` | `{message}`（见缺口⑥） |

`Shop` 含返利字段：`reward_rate_self`(默认0.03)、`reward_rate_level1`(0.10)、`reward_rate_level2`(0.04)、`reward_ceiling`(0.50)、`reward_exclude_categories`(jsonb 数组)。
`Product.status`：`1=上架, 0=下架, 2=售罄`。

---

## 5. 后端 API 缺口（需先补，否则前端做不到）

> 这几项是写前端前要确认/修复的依赖。建议作为 spec 的第一批后端任务。

**① 店铺返利配置无法保存** — `UpdateShopRequest`（`api/handler/shop.go`）不含返利字段。
→ 给 `UpdateShopRequest` 加 `reward_rate_self/level1/level2/ceiling/reward_exclude_categories`，并允许写入（注意 0 值合法，不能用 `if x > 0` 守卫，改用指针或 `*float64`）。

**② 菜品「下架」(status=0) 与清空字段做不到** — `UpdateProduct` 用 `if req.Status > 0`、`if req.Name != ""` 等部分更新，无法把 status 设为 0，也无法清空描述/图片。
→ 改用 `*` 指针字段或 `map` 显式判定 `status` 是否传入。最小改动：单独支持 status 写 0。

**③ 桌码接口未鉴权且不在 merchant 路径** — `POST/GET /api/shops/:id/qrcodes` 公开（代码有 TODO）。
→ **已定：方案 A** — 新增 `GET/POST /api/merchant/shops/:id/qrcodes`，校验 shop 归属当前 merchant；前端调这组。旧公开接口保留（小程序扫码用 `/scan`，不受影响）。

**④ 看板无趋势/区间数据** — 仅单日 `stats` + 累计 `dashboard`。
→ **已定：v1 只做卡片** — 「累计卡片 + 单日卡片 + 日期选择器」，不做折线图、不引 ECharts。

**⑤ 无图片上传接口** — `Product.Image` 是 URL 字符串。
→ **已定：本期做上传**。新增 `POST /api/merchant/upload`（multipart，校验商家登录、类型 jpg/png/webp、大小上限 5MB），返回 `{url}`，写回 `Product.Image`。
→ **已定：存储用 Cloudflare R2**（见 §10），起步即上 R2，不做本地磁盘过渡。

**⑥ 订单运营台后端（新模块）** — `merchant_order.go` 现仅有 `GetMerchantOrders`（裸 `Order`，固定 50 条、无分页、不含配送明细）+ `CreateMerchantOrder`。需补：
- **a) 出餐字段**：`Order` 加 `PreparedAt *time.Time`（出餐时间；`已出餐` = 非空）。**不**新增 `Status` 枚举值。→ **迁移前** grep 所有 `Status` 读取点，确认无处把 `Status==3` 当"已出餐"语义。
- **b) 列表扩展**：`GetMerchantOrders` LEFT JOIN `order_deliveries`，返回 `order_type / status / paid_at / prepared_at` + `delivery{shansong_status,...}`；增加 `status`/`type` 筛选 + 分页（替换现固定 50 条上限）。**未支付(Status=1)也返回**（商家需知"是否已支付"，用支付状态徽标区分）。
- **c) 出餐**：`POST /api/merchant/orders/:id/prepare` → 置 `PreparedAt`，校验 order 所属 shop 归当前 merchant。
- **d) 重新派单**：`POST /api/merchant/orders/:id/redispatch` → 调 `services.DispatchShansong(orderID)`，仅限 `ShansongStatus ∈ {-1 派单失败, 60 已取消}`，校验归属。→ **开工前**读 `services/shansong_dispatch.go` 确认其重新派单时会**重新询价**（旧 `ShansongQuoteNo` 可能已失效）。
- **e) 改状态**：`PUT /api/merchant/orders/:id/status` → 手动改 `Order.Status`，校验归属。
→ 4 个新/改 handler 均补 Go `httptest`（沿用 §7 风格）。
→ **产品风险（非技术）**：堂食出餐依赖厨房实时在后台点【出餐】，否则字段形同虚设——开工前向真实商家确认出餐履约流程。

---

## 6. Code Style

- Vue SFC `<script setup>`，组件 PascalCase，文件名与组件名一致。
- API 层只返回 `response.data`，错误经 axios 拦截器统一 `ElMessage` 提示。
- 不引入 TS、不引入 Tailwind（与仓库现状一致，避免额外工具链）。
- 金额/比例展示：返利比例用百分比输入（UI 显示 3%，存 0.03），由前端换算。
- 不做过度抽象：单次使用的逻辑不抽组件/composable（CLAUDE.md §2）。

---

## 7. Testing Strategy

- **手动联调为主**（中后台 CRUD，e2e 成本高收益低）：每个模块以「能登录→能增删改查→刷新后状态正确」为验收。
- **关键纯函数单测**（Vitest）：百分比↔小数换算、axios 拦截器 token 注入、401 跳转。
- **后端缺口修复**：用 Go `httptest` 给改动的 handler（①②③）补最小单测，沿用现有 `api/handler/handler_test.go` 风格。
- 不为 UI 组件写快照测试。

---

## 8. Boundaries

**总是做**
- 所有商家接口带 `Authorization` 头；401 自动登出跳登录。
- 改后端 handler 时配套单测，保持 `go test ./...` 绿。
- 前端改动局限在新建的 `admin/` 目录，不动小程序 `frontend/`。

**先问再做**
- 修改后端任何已有 handler 的现有字段/行为（除 §5 列出的缺口外）。
- 引入 §2 之外的新依赖（图表库、上传 SDK 等）。
- 桌码鉴权方案 A/B（缺口③）、图片上传方案（缺口⑤）。

**绝不做**
- 不覆盖根目录 `SPEC.md`、不改 `frontend/`（小程序）业务逻辑。
- 不在前端硬编码 token/密钥；后端 secret 仍走 config/env。
- 不删除现有未用代码（除非本次改动产生的孤儿）。

---

## 9. 建议构建顺序（每步可验证）

1. 后端补缺口 ①②③⑤⑥ + 单测 → `go test ./...` 通过：
   - ① 店铺返利字段可写（指针/显式 map）
   - ② 菜品 status 可设 0（下架）
   - ③ 新增 `/api/merchant/shops/:id/qrcodes`（GET/POST，校验归属）
   - ⑤ 新增 `POST /api/merchant/upload`（multipart → `{url}`）
   - ⑥ `Order.PreparedAt` 迁移 + 列表 JOIN/筛选/分页 + prepare/redispatch/status 三接口
2. `admin/` Vite 脚手架 + axios 封装 + 登录页 → 能登录拿到 token、刷新保持。
3. Layout + 店铺切换器 + 路由守卫 → 未登录跳登录。
4. 菜品管理（列表/新增/编辑/上下架/删除 + 图片上传）→ CRUD 闭环。
5. 店铺设置 + 返利配置 → 保存后刷新值正确。
6. 看板卡片 + 日期选择 → 数字与后端一致。
7. 桌码生成/列表 → 能出图、能看列表。
8. 订单运营台 → 默认"待处理"队列（堂食已支付未出餐 + 外卖派单失败/卡太久）；堂食可【出餐】、外卖可【重新派单】/【改状态】；tab 切进行中/已完成/全部；按店铺/日期筛选；未支付订单带支付状态徽标。
```
每步 verify：手动操作一遍 + 控制台无报错 + 刷新后状态正确。
```

---

## 10. 图片上传存储方案：Cloudflare R2（已定）

**选型理由**：R2 出口流量永久免费（菜品图被小程序反复加载不计费），10GB 免费存储够起步，S3 兼容可直接用 `aws-sdk-go-v2`。

**后端实现**（`POST /api/merchant/upload`）：
- 用 `aws-sdk-go-v2` + `aws-sdk-go-v2/service/s3`，`BaseEndpoint` 指向 R2 endpoint，region 设 `auto`。
- 服务端转存：接收 multipart → 校验类型(jpg/png/webp)/大小(≤5MB) → 生成随机对象名（如 `products/<uuid>.<ext>`）→ `PutObject` 到 bucket → 返回 `{url}`（公开域名 + key 拼成）。
- `Image` 存完整公网 URL。

**所需 env**（走 config/env，同微信支付密钥的管理方式）：
| 变量 | 说明 |
|------|------|
| `R2_ACCOUNT_ID` | Cloudflare 账户 ID |
| `R2_ACCESS_KEY_ID` | R2 API Token 的 Access Key |
| `R2_SECRET_ACCESS_KEY` | R2 API Token 的 Secret |
| `R2_BUCKET` | 存储桶名 |
| `R2_PUBLIC_BASE` | 公开访问域名，**已定 `https://bestluckbox.com`**（自定义域，走 Cloudflare CDN）；`Image` = `R2_PUBLIC_BASE` + `/` + objectKey |

> endpoint 形如 `https://<R2_ACCOUNT_ID>.r2.cloudflarestorage.com`（S3 API），与 `R2_PUBLIC_BASE`（读取域名）是两个不同地址，别混。
>
> **前置依赖（用户操作）**：开工前需在 Cloudflare 建好 bucket、开启公开访问、生成 API Token，拿到上表 5 个值。步骤见交付时附带的清单。
