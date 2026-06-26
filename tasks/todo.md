# Todo：闪送集成层对齐真实 merchants/v5 协议（修正）

详见 `tasks/plan.md`。一任务一提交，TDD（RED→GREEN→回归）。
背景：`/iss-open-cli` 验证发现 T2 对接层按猜测实现，与真实闪送不符；本计划只修正对接层，业务口径不变。
基线：绝不 git add `frontend/config.js` / `.claude/` / `specs/shansong-delivery.md`；密钥仅 env。
注：本次会更新 T2/T3/T4/T5/T6 既有测试——它们此前锁定错误协议，对齐真实契约是修正的一部分。

已确认决策：新增 Shop.City（商家维护）；/quote 生成 order_no 写入凭证、下单复用；保留回调 webhook、改真实验签+状态码。

## 协议层
- [ ] **FT1** 传输 form-urlencoded + 真实签名（含 shopId）+ 系统参数 + merchants/v5 端点；config 加 ShopID — M
- [ ] **FT2** orderCalculate 请求结构（cityName/sender/receiverList，字符串坐标，goodType=6）+ orderPlace 仅 issOrderNo + 状态码 20/30/40/50/60 — M

## 数据 + 贯通
- [ ] **FT3** Shop 增 City 字段（model + UpdateShop + DTO）— S
- [ ] **FT4** /quote 生成 order_no + 城市/收发件映射；token 内嵌 order_no+issOrderNo；CreateOrder 复用 order_no — M

## 派单 + 回调
- [ ] **FT5** DispatchShansong→orderPlace，初始状态改 20（原 60=已取消，必错）— S
- [ ] **FT6** 回调验签改真实规则 + 解析 issOrderNo/orderStatus — M

## Checkpoint 联调
- [ ] 三套测试全绿（go / jest / vitest）
- [ ] 真实凭据跑 orderCalculate 确认运费字段名（回填 FT2 CALIBRATION）
- [ ] 校准回调入站格式（FT6 CALIBRATION）
- [ ] 真机全链路：选址→报价→支付→orderPlace→回调→我的订单
- [ ] 仓库无真实密钥；重出 GO/NO-GO

## 不在本期
- 取消/退款 UI（abortOrder 客户端已留方法）
- orderInfo 轮询（本期靠回调；abortOrder/orderInfo 留作后续）
