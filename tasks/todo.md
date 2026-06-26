# Todo：闪送集成层对齐真实 merchants/v5 协议（修正）

详见 `tasks/plan.md`。一任务一提交，TDD（RED→GREEN→回归）。
背景：`/iss-open-cli` 验证发现 T2 对接层按猜测实现，与真实闪送不符；本计划只修正对接层，业务口径不变。
基线：绝不 git add `frontend/config.js` / `.claude/` / `specs/shansong-delivery.md`；密钥仅 env。
注：本次会更新 T2/T3/T4/T5/T6 既有测试——它们此前锁定错误协议，对齐真实契约是修正的一部分。

已确认决策：新增 Shop.City（商家维护）；/quote 生成 order_no 写入凭证、下单复用；保留回调 webhook、改真实验签+状态码。

## 协议层
- [x] **FT1** 传输 form-urlencoded + 真实签名（含 shopId）+ 系统参数 + merchants/v5 端点；config 加 ShopID — M ✅
- [x] **FT2** orderCalculate 请求结构（cityName/sender/receiverList，字符串坐标，goodType=6）+ orderPlace 仅 issOrderNo + 状态码 20/30/40/50/60 — M ✅

## 数据 + 贯通
- [x] **FT3** Shop 增 City 字段（model + UpdateShop + DTO）— S ✅
- [x] **FT4** /quote 生成 order_no + 城市/收发件映射；token 内嵌 order_no+issOrderNo；CreateOrder 复用 order_no — M ✅

## 派单 + 回调
- [x] **FT5** DispatchShansong→orderPlace，初始状态改 20（原 60=已取消，必错）— S ✅
- [x] **FT6** 回调验签改真实规则 + 解析 issOrderNo/orderStatus — M ✅

## Checkpoint 联调
- [x] 三套测试全绿（go / jest 94 / vitest 21）✅
- [x] 仓库无真实密钥（secrets 扫描通过；config.yaml.example 仅占位）✅
- [x] 真实凭据跑 orderCalculate：运费字段=totalFeeAfterSave/totalAmount（单位**分**）；已修 FT2 ✅
- [x] 本地 ShansongClient 实测 test 环境 orderCalculate 成功（fee=38.64 元，issOrderNo 返回）✅
- [x] orderPlace 实际下单成功（status 200，订单创建）✅
- [x] orderStatus 枚举实测确认：**20=派单中**（与 FT2 状态文案 + FT5 派单初始状态一致；旧 60 bug 确认已修）✅
- [x] abortOrder 取消成功（status 200，已清理测试单）✅
- [x] orderInfo 字段 `orderStatus` 与 FT6 回调解析一致（佐证）✅
- [ ] 回调入站完整 wrapper（FT6）：字段 orderStatus 已佐证；form/data 完整格式待真实回调推送确认
- [ ] 小程序真机全链路 + 重出 GO/NO-GO
- 备注：orderInfo 需 thirdOrderNo（合作伙伴单号），非本期范围（用回调而非轮询）

## 不在本期
- 取消/退款 UI（abortOrder 客户端已留方法）
- orderInfo 轮询（本期靠回调；abortOrder/orderInfo 留作后续）
