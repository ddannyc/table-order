# Todo：闪送外卖配送对接

详见 `tasks/plan.md`。一任务一提交，TDD（RED→GREEN→回归）。
基线：绝不 git add `frontend/config.js` / `.claude/` / `specs/shansong-delivery.md`；闪送真实密钥绝不入库（仅 env / gitignore 的 config.yaml）。

已确认决策：目的地坐标用 `wx.getLocation`（无后端地理编码）；取消/退款本期不做；配送状态展示纳入本期；闪送凭据就绪、本期含联调。

## P1 后端基础 + 报价
- [x] **T1** OrderDelivery 模型（+收件坐标）+ 迁移注册 + ShansongConfig + yaml 占位 — S ✅
- [x] **T2** 闪送 service：签名 + CalculatePrice + CreateOrder + 状态映射（mock 单测）— M ✅
- [x] **T3** `POST /api/delivery/quote`：报价透传 + 后端签名报价凭证 + 坐标缺失兜底 — M ✅

## P2 下单与履约
- [ ] **T4** CreateOrder delivery 分支：验凭证落 OrderDelivery，实付含运费，返利按菜品额 — M
- [ ] **T5** 支付成功后 `DispatchShansong`（wechatpay_notify + 零元分支钩子，仅 delivery）— M
- [ ] **T6** `POST /api/shansong/callback`：验签 + 更新状态 + 幂等 + 返回 `{"status":200}` — M

## P3 前端 + 状态展示
- [ ] **T7** order-confirm：chooseAddress + getLocation + 报价 + 应付含运费 + payload 扩展；app.json 加 getLocation — M
- [ ] **T8** GetOrders/GetOrder 回传 delivery + 我的订单展示配送状态 — M

## P4 联调 Checkpoint
- [ ] 三套测试全绿（go / jest / vitest）
- [ ] 真机：chooseAddress+getLocation 授权（mp 后台开通 getLocation + 隐私指引）
- [ ] 测试环境全链路：选址→报价→支付→派单→回调→我的订单可见
- [ ] 堂食回归通过；运费不进返利/福利金
- [ ] 仓库无闪送真实密钥；重出 GO/NO-GO

## 不做（本期）
- 取消/退款（闪送取消 + 退运费）— 留后续
- 后端用户地址簿 — 仅本地缓存
- 多门店就近选店 — 维持单店 stub
