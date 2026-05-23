# Task List: 点餐功能扩展

## Phase 1: Foundation

### Task 1: DB migration + Product model
**Acceptance:**
- [ ] products 表创建，字段：id, shop_id, name, price, description, image, category, status, created_at, updated_at
- [ ] order_items 表创建，字段：id, order_id, product_id, product_name, price, quantity, subtotal
- [ ] Go models: product.go, order_item.go

**Verification:**
- `PGPASSWORD=postgres psql -d table_order -c "\d products"` - 表存在
- `PGPASSWORD=postgres psql -d table_order -c "\d order_items"` - 表存在
- `cd backend && go build ./...` - 编译通过

**Files:**
- `backend/models/product.go` (new)
- `backend/models/order_item.go` (new)
- SQL migration (auto via GORM)

---

### Task 2: Seed test products
**Acceptance:**
- [ ] 店铺ID=1插入5个测试菜品（涵盖不同分类和状态）
- [ ] 菜品状态含：上架(1)、售罄(2)

**Verification:**
- `PGPASSWORD=postgres psql -d table_order -c "SELECT name, price, category, status FROM products;"` - 显示5条

**Files:**
- SQL insert (手动执行)

---

## Phase 2: Core User Flow

### Task 3: Product API (用户端)
**Acceptance:**
- [ ] `GET /api/shops/:id/products` 返回店铺所有上架(status=1)菜品
- [ ] 按 category 分组返回

**Verification:**
- `curl --noproxy '*' http://localhost:8080/api/shops/1/products` - 返回菜品数组

**Files:**
- `backend/api/handler/product.go` (new)

---

### Task 4: Frontend product API + scan.vue 改造
**Acceptance:**
- [ ] `frontend/src/api/product.js` 包含 `getShopProducts(shopId)`
- [ ] scan.vue 调用 API 显示菜品列表（分类 + 名称 + 价格 + 图片）
- [ ] 点击菜品弹出数量选择，加购后存到本地 storage
- [ ] 底部购物车栏显示已选商品数量和总价

**Verification:**
- 启动 `npm run dev:h5`，扫码进入页面显示菜品列表

**Files:**
- `frontend/src/api/product.js` (new)
- `frontend/src/pages/scan/scan.vue` (改造)

---

### Task 5: Order API 扩展 (含 OrderItems)
**Acceptance:**
- [ ] CreateOrderRequest 新增 `items []OrderItemRequest`
- [ ] 后端重算订单金额 = sum(item.price * item.quantity)
- [ ] 订单创建时同时创建 order_items 记录
- [ ] 前端传 product_id + quantity，后端取 price

**Verification:**
- 下单成功后 `PGPASSWORD=postgres psql -d table_order -c "SELECT * FROM order_items;"` 有记录

**Files:**
- `backend/api/handler/order.go` (改造)
- `backend/models/order_item.go` (已创建)

---

### Task 6: Balance pay API
**Acceptance:**
- [ ] `POST /api/orders/:id/pay` - 余额支付
- [ ] 扣款前检查余额是否充足
- [ ] 事务：扣余额 + 更新订单status=2 + 创建wallet_log
- [ ] 余额不足返回 error

**Verification:**
- 余额充足时支付成功，订单status=2
- 余额不足时返回 "balance not enough"

**Files:**
- `backend/api/handler/order.go` (改造)

---

### Task 7: cart.vue + order-confirm.vue
**Acceptance:**
- [ ] cart.vue: 显示购物车商品列表，数量+/-，删除，单项小计，总价，预计返利
- [ ] order-confirm.vue: 订单摘要（店铺/桌号/商品列表/金额），余额支付按钮
- [ ] 支付成功后跳转到订单完成页

**Verification:**
- 完整流程：加购 → 购物车 → 确认订单 → 支付 → 显示订单完成

**Files:**
- `frontend/src/pages/cart/cart.vue` (new)
- `frontend/src/pages/order-confirm/order-confirm.vue` (new)

---

## Phase 3: Merchant CRUD

### Task 8: Product CRUD (商户端)
**Acceptance:**
- [ ] `POST /api/merchant/products` - 创建菜品（需商户认证）
- [ ] `GET /api/merchant/products` - 返回商户自己店铺的菜品
- [ ] `PUT /api/merchant/products/:id` - 更新菜品（含 status 改为售罄）
- [ ] `DELETE /api/merchant/products/:id` - 删除菜品（软删或硬删）

**Verification:**
- 商户登录后可以 CRUD 自己的菜品

**Files:**
- `backend/api/handler/product.go` (改造，商户端接口)

---

## Checkpoints

### Checkpoint 1 (After Task 1-2)
- [ ] products, order_items 表存在
- [ ] go build 通过
- [ ] 测试数据就绪

### Checkpoint 2 (After Task 3-7)
- [ ] 用户完整流程可跑通
- [ ] 余额支付正常

### Checkpoint 3 (After Task 8)
- [ ] 商户可管理菜品
- [ ] 所有功能测试通过