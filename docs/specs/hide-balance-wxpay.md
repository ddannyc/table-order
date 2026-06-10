# Spec: 隐藏余额支付，默认调起微信支付

## Objective

隐藏前端余额支付选项，下单后直接调起微信支付（JSAPI/小程序支付）。

**用户故事：** 顾客扫码点餐 → 下单确认页不再展示"余额支付"选项 → 点击"确认支付"直接调起微信支付收银台 → 支付成功后跳转订单页。

**变更范围：**
- 前端：order-confirm 页面移除"支付方式"卡片，按钮改为"微信支付"，调用 `wx.requestPayment`
- 后端：下单接口从"创建订单+扣余额"改为"创建待支付订单+调微信统一下单+返回支付参数"

## 前提条件 / 阻塞项

⚠️ **微信支付需要以下配置，当前项目没有：**
1. 微信支付商户号 (MchID)
2. API v3 密钥 (或 API v2 key)
3. 商户证书（API v3 需要）
4. 微信支付回调地址（用于支付结果通知 `notify_url`）

**这些必须在实施前获取。** 没有商户号，后端无法调用统一下单 API。

## Tech Stack

- 前端：微信小程序原生框架（WeUI 组件库）
- 后端：Go + Gin + GORM
- 支付：微信支付 JSAPI（小程序支付）

## Commands

```
Build:  cd frontend && npm run build (或微信开发者工具编译)
Test:   cd backend && go test ./...
Lint:   cd backend && go vet ./...
Dev:    cd backend && go run main.go
```

## Project Structure

```
frontend/pages/order-confirm/   ← 主要改动：支付方式 UI + 支付逻辑
frontend/api/index.js            ← 新增 createWxPayOrder API
backend/api/handler/order.go     ← CreateOrder 重构：不扣余额，调微信统一下单
backend/services/wechat.go       ← 新增微信支付相关函数
backend/config/config.go         ← 新增微信支付配置项
backend/api/router/router.go     ← 新增支付回调路由
```

## Code Style

遵循项目现有风格。前端 ES5 `require/module.exports`，后端 Go 标准项目结构。

## Testing Strategy

- 后端：`go test ./...` 确保现有测试通过
- 前端：微信开发者工具手动测试支付流程
- 支付回调：用微信支付沙箱环境测试

## Boundaries

- **Always:** 保持现有订单模型兼容，不破坏已有订单查询功能
- **Ask first:** 微信支付商户号配置、支付回调域名
- **Never:** 不删除余额相关数据库字段（用户余额、wallet_log 等，未来可能恢复）

## Success Criteria

1. ✅ order-confirm 页面不展示"余额支付"选项
2. ✅ 点击"确认支付"按钮，前端调用后端下单接口
3. ✅ 后端创建待支付订单（status=1），调用微信统一下单获取 prepay_id
4. ✅ 后端返回支付参数给前端，前端调用 `wx.requestPayment` 拉起微信支付
5. ✅ 支付成功后，微信回调更新订单状态为已支付
6. ✅ 支付成功后，前端跳转到订单/个人页
7. ✅ 现有余额查询、钱包流水等功能不受影响

## Open Questions

1. **微信支付商户号、API密钥是否已就绪？** 如果没有，需要先申请。
2. **支付回调域名是否已配置？** 微信支付要求回调 URL 为已备案域名。
3. **余额充值入口是否保留？** 当前需求只涉及支付端。如果余额不再用于支付，充值功能是否需要调整？
4. **福利金（reward_balance）抵扣逻辑如何处理？** 微信支付需要传最终金额，福利金抵扣是否需要保留？
