# Implementation Plan: 点餐功能扩展

## Overview

实现扫码 → 点餐页 → 加购 → 下单 → 支付完整流程。商户创建菜品管理，用户扫码进入店铺页选择菜品，购物车本地持久化，余额支付下单。

## Architecture Decisions

1. **垂直切片**：按功能路径实现，而非按层实现。先完成后端订单+支付完整路径，再实现商户菜品管理。
2. **前端购物车**：使用 uni-app storage 本地持久化，key = `cart_${shopId}`
3. **后端金额计算**：订单金额后端计算，前端只传商品ID和数量，防止篡改
4. **无库存管理**：通过菜品 `status=2` (售罄) 控制，而非库存数量

## Dependency Graph

```
DB schema (products, order_items)
    │
    ├── Go models (product.go, order_item.go)
    │       │
    │       ├── Product API (merchant + user endpoints)
    │       │       │
    │       │       ├── Order API (with items)
    │       │       │
    │       │       ├── Balance pay API
    │       │       │
    │       │       └── Frontend API clients
    │       │
    │       └── Frontend pages
    │
    └── Seed data (test products)
```

## Task List

### Phase 1: Foundation (DB + Models)

**Task 1: Database migration + Product model**
- 创建 products 表（含 shop_id, name, price, category, status）
- 创建 order_items 表（含 order_id, product_id, price, quantity, subtotal）
- 写入 Go models

**Task 2: Seed test products**
- 为测试店铺插入 3-5 个菜品

**Checkpoint 1: Foundation**
- [ ] `psql -c "\d products"` 正常
- [ ] `psql -c "\d order_items"` 正常
- [ ] `go build ./...` 通过

---

### Phase 2: Core User Flow (点餐 → 支付)

**Task 3: Product API (用户端)**
- `GET /api/shops/:id/products` - 获取店铺菜品列表(上架)
- 无需认证

**Task 4: Frontend product API + scan.vue 改造**
- `GET /api/shops/:id/products` 调用
- scan.vue 显示菜品分类列表
- 点击加购存入本地 storage

**Task 5: Order API 扩展 (含 OrderItems)**
- 修改 CreateOrderRequest 接受 items 数组
- 创建订单时同时创建 order_items
- 计算订单总金额 = sum(item.price * item.quantity)

**Task 6: Balance pay API**
- `POST /api/orders/:id/pay` - 余额扣款
- 扣款前检查余额是否充足
- 事务保证：扣余额 + 更新订单状态 + 创建 wallet_log

**Task 7: order-confirm.vue + cart.vue**
- cart.vue: 购物车列表，修改数量/删除，显示总价
- order-confirm.vue: 订单摘要，选择支付方式，余额支付

**Checkpoint 2: Core User Flow**
- [ ] 扫码进入店铺看到菜品列表
- [ ] 加购后购物车有记录
- [ ] 下单成功并扣减余额
- [ ] 订单详情含商品明细

---

### Phase 3: Merchant CRUD

**Task 8: Product CRUD (商户端)**
- `POST /api/merchant/products` - 创建菜品
- `GET /api/merchant/products` - 商户菜品列表
- `PUT /api/merchant/products/:id` - 更新菜品（含设为售罄）
- `DELETE /api/merchant/products/:id` - 删除菜品

**Checkpoint 3: Complete**
- [ ] 商户可创建/编辑/删除菜品
- [ ] 商户可将菜品设为售罄
- [ ] 所有测试通过

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| 余额支付并发 | 中 | 事务 + 乐观锁检查余额 |
| 前端购物车丢失 | 低 | 本地 storage + 每次操作回写 |
| 订单金额前端伪造 | 高 | 后端重算金额，不信任前端 |

## Open Questions

None - 所有问题已在 spec 阶段澄清。