# 实施计划：闪送集成层对齐真实 merchants/v5 协议（修正）

> 接 `/iss-open-cli` 验证结论：T1–T8 的**业务流程/架构正确**，但 T2 闪送**接口对接层**（CALIBRATION 点）几乎全部按猜测实现，与真实 `openapi/merchants/v5` 不符。本计划只修正对接层，使其可与真实闪送互通。**业务口径（运费不进返利、HMAC 凭证、getLocation 坐标、幂等、异步派单）保持不变。**

## 背景：已确认的真实协议（来源 iss-open-cli 文档 + 示例）
- 传输：`POST application/x-www-form-urlencoded`；系统参数 `clientId`、`shopId`、`timestamp`、`sign`；业务参数 `data`（紧凑 JSON 字符串）。
- 签名：`MD5(appSecret + "clientId" + clientId + "data" + data + "shopId" + shopId + "timestamp" + timestamp)` 转大写；`data` 为空时省略 `"data"+data`。
- 端点：`/openapi/merchants/v5/orderCalculate | orderPlace | orderInfo | abortOrder`。
- orderCalculate 业务体：`{cityName, sender:{fromAddress,fromSenderName,fromMobile,fromLatitude,fromLongitude}, receiverList:[{orderNo,toAddress,toLatitude,toLongitude,toReceiverName,toMobile}]}`，**经纬度为字符串**；返回 `orderNumber`（后续作 `issOrderNo`）。
- orderPlace 业务体：**仅** `{issOrderNo}`。
- 订单状态枚举：`20 派单中 / 30 待取货 / 40 闪送中 / 50 已完成 / 60 已取消`。
- 物品类型：餐饮=6（本期用 6，CLI 默认 10）。

## 本期已确认决策（人工）
1. **cityName**：新增 `Shop.City` 字段，商家后台维护（同 Latitude/Longitude 模式）。
2. **第三方单号**：`/quote` 生成 `order_no`，写入 HMAC 凭证；CreateOrder 复用为订单 `order_no`，并作 `receiverList[].orderNo`（= 闪送 `thirdOrderNo`，可与 orderInfo 交叉查询）。
3. **状态更新**：保留回调 webhook（T6），仅把验签改为真实签名规则、状态码改真实枚举。

## 假设（如有异议请纠正）
- 后端继续自行发起 HTTP 调用，**不依赖本地 `iss-open-cli`**（CLI 仅为学习/联调工具）。
- 回调入站报文沿用同一协议（form-urlencoded + clientId/shopId/timestamp/sign/data，`data` 含 `issOrderNo` + `orderStatus`）——**精确格式以闪送回调文档为准，标 CALIBRATION**。
- orderCalculate 响应的**总运费字段名**未能从 CLI 文档确认（无真实凭据）——按最可能字段实现并标 CALIBRATION，联调时以真实响应校准。

## 基线纪律
绝不 `git add` `frontend/config.js`、`.claude/`、`specs/shansong-delivery.md`；闪送真实密钥仅经 env / gitignore 的 `config.yaml`。**本次会更新 T2/T3/T4/T5/T6 既有测试**——这些测试此前锁定的是错误协议，更新它们对齐真实契约是修正的一部分，不是“改测试迁就实现”。

## 依赖 / 执行序
FT1（传输+签名+config+端点）→ FT2（请求/响应结构+状态码）→ FT3（Shop.City）→ FT4（/quote 单号+城市+收发件映射 & CreateOrder 复用单号）→ FT5（派单 orderPlace + 初始状态）→ FT6（回调验签+字段）→ Checkpoint。
- FT3 独立，可提前；置于 FT4 前（FT4 依赖 City）。
- FT4 依赖 FT2（CalculatePrice 新结构）+ FT3。FT5 依赖 FT2。FT6 依赖 FT1（签名）。

---

## 任务列表

### FT1 —【协议层】传输 + 签名 + 系统参数 + 端点对齐真实
**做法：** `ShansongConfig` 增 `ShopID`（config + `SHANSONG_SHOP_ID` env + yaml 占位 + `InitShansongClient` 签名加 shopID）。`signShansong` 改真实拼接规则（含 shopId，alpha 序 clientId/data/shopId/timestamp，appSecret 前缀，data 空则省略）。`post` 改 `application/x-www-form-urlencoded`，表单字段 `clientId/shopId/timestamp/sign/data`（data 为紧凑 JSON）。端点常量改 `/openapi/merchants/v5/{orderCalculate,orderPlace,orderInfo,abortOrder}`。响应信封 `{status,msg,data}` 解析不变。
**Acceptance:**
- [ ] 已知输入 → 期望签名（真实规则；含 shopId）。
- [ ] 出站请求为 form-urlencoded，含 clientId/shopId/timestamp/sign/data；sign 覆盖 data。
- [ ] base_url + merchants/v5 端点正确。
**Verification:** [ ] 重写 `services/shansong_test.go` 签名/请求构造用例为真实契约；`go test ./...` 全绿。
**Files:** `backend/services/shansong.go`、`backend/config/config.go`、`backend/config/config.yaml.example`、`backend/main.go`、`backend/services/shansong_test.go`
**Scope:** M

### FT2 —【业务体】orderCalculate/orderPlace 请求响应 + 状态码
**做法：** `QuoteRequest` 增 `CityName`、`SenderName`、`SenderMobile`、`ThirdOrderNo`。`CalculatePrice` 构造真实 `{cityName, sender{from*}, receiverList:[{orderNo,to*}]}`，**经纬度转字符串**，`goodType=6`、`weight=1`、`appointType=0`；响应解析 `orderNumber` + 运费字段（CALIBRATION）。`CreateOrder`（→orderPlace）业务体改**仅** `{issOrderNo}`（= QuoteToken）。`shansongStatusLabels` 改 `20/30/40/50/60`，默认兜底文案。
**Acceptance:**
- [ ] orderCalculate 请求体结构/字段名/字符串坐标与真实一致；goodType=6。
- [ ] orderPlace 仅传 issOrderNo。
- [ ] 状态文案：20→派单中…60→已取消；未知码兜底非空。
**Verification:** [ ] 更新 service 用例（mock HTTP 校验请求体 + 解析）；`go test ./...` 全绿。
**Files:** `backend/services/shansong.go`、`backend/services/shansong_test.go`
**Scope:** M

### FT3 —【数据】Shop 增 City 字段
**做法：** `Shop` 增 `City string`；`UpdateShop` 接受 `city`；`publicShopDTO` 可含 city（展示无害）。迁移随 AutoMigrate（仅增列）。
**Acceptance:**
- [ ] `Shop.City` 落库；`UpdateShop` 可写入。
**Verification:** [ ] 补/改 handler 测试断言 city 可写读；`go test ./...` 全绿。
**Files:** `backend/models/shop.go`、`backend/api/handler/shop.go`、相关 `_test.go`
**Scope:** S

### FT4 —【贯通】/quote 生成单号 + 城市/收发件映射；CreateOrder 复用单号
**做法：** `quoteClaims` 增 `OrderNo`。`DeliveryQuote`：生成 `order_no`（时间戳+uuid，沿用现格式），取 `shop.City` 缺失→400；调 `CalculatePrice` 传 cityName + sender（shop.Name/Phone/Address/坐标）+ receiver（请求收件信息）+ `ThirdOrderNo=order_no`；token 内嵌 `OrderNo` 与 `ShansongQuote(=issOrderNo)`。`CreateOrder`：delivery 分支用 `claims.OrderNo` 作订单 `OrderNo`（不再新生成），其余不变。
**Acceptance:**
- [ ] /quote 在 shop.City 缺失时 400；正常返回 fee + token（token 含 order_no + issOrderNo）。
- [ ] delivery 订单的 `order_no` == 凭证内 order_no；`OrderDelivery.ShansongQuoteNo` == issOrderNo。
- [ ] 堂食 order_no 生成不变。
**Verification:** [ ] 更新 `delivery_test.go` / `order_create_delivery_test.go`；`go test ./...` 全绿。
**Files:** `backend/api/handler/delivery.go`、`backend/api/handler/order.go`、相关 `_test.go`
**Scope:** M

### FT5 —【派单】DispatchShansong orderPlace + 初始状态
**做法：** `DispatchShansong` 调 orderPlace（`issOrderNo=od.ShansongQuoteNo`），成功落 `ShansongOrderNo`（闪送返回单号，可能与 issOrderNo 相同/不同——以响应为准）+ 初始状态改 **20（派单中）**（原 60 在真实枚举=已取消，必错）。其余（幂等、仅 delivery、失败仅记日志）不变。
**Acceptance:**
- [ ] 派单成功落单号 + 初始状态 20。
- [ ] 失败不阻塞；非 delivery no-op；幂等。
**Verification:** [ ] 更新 `shansong_dispatch_test.go` 断言初始状态 20；`go test ./...` 全绿。
**Files:** `backend/services/shansong_dispatch.go`、`backend/services/shansong_dispatch_test.go`
**Scope:** S

### FT6 —【回调】验签真实规则 + 字段对齐
**做法：** `VerifyCallback` 已用 `CallbackSign`→`signShansong`，FT1 改对后自动对齐；回调入站改按 form/`data` 解析 `issOrderNo` + `orderStatus`（替换 orderNumber/status 命名），按 `shansong_order_no`/issOrderNo 匹配更新；返回 `{"status":200}` 不变。**入站精确格式标 CALIBRATION。**
**Acceptance:**
- [ ] 真实签名规则验签通过/失败。
- [ ] 解析 issOrderNo + orderStatus 更新状态；幂等；返回 `{"status":200}`。
**Verification:** [ ] 更新 `shansong_notify_test.go` 为真实字段/签名；`go test ./...` 全绿。
**Files:** `backend/api/handler/shansong_notify.go`、`backend/api/handler/shansong_notify_test.go`
**Scope:** M

### Checkpoint（联调）
- [ ] 三套测试全绿：`go test ./...`、`jest`、`vitest`。
- [ ] 配置真实测试凭据（`~/.iss-open-cli/configs/config.yaml`）跑一次 `orderCalculate`，**确认运费字段名**与响应结构，回填 FT2 的 CALIBRATION。
- [ ] 校准回调入站报文格式（FT6 CALIBRATION）以闪送回调文档/真实回调为准。
- [ ] 真机：选址→报价→支付→orderPlace 派单→回调/状态→我的订单。
- [ ] 仓库无闪送真实密钥；重出 GO/NO-GO。

---

## 风险
| 风险 | 缓解 |
|------|------|
| 运费字段名/回调格式仍未实测确认 | 标 CALIBRATION + Checkpoint 用真实凭据/文档校准；逻辑解耦便于一处改 |
| Shop.City 历史数据为空 → 外卖报价 400 | /quote 明确 400 提示；商家后台补 City（同坐标维护流程） |
| 闪送返回单号 vs issOrderNo 语义 | 以真实响应为准；派单落 ShansongOrderNo，回调/查询用 issOrderNo |
| 改动既有测试面较大 | 逐 FT 提交；每步全绿；既有业务测试（返利/堂食）不应回归 |

## 回滚
- 分支 `feat/shansong-delivery` 未 push/未合并/未部署：回滚 = 不合并。
- AutoMigrate 仅增列（Shop.City），幂等无害。
