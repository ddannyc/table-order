# Todo: 堂食/外卖双模式 + 菜品页改版 + SKU + 全站换肤

详见 `tasks/plan.md`。每阶段后必须通过检查点（含人工确认）才进入下一阶段。
设计规范唯一真源：`DESIGN.md`。

## Phase 0 — 品牌换肤地基（横切，低风险）
- [x] **T1** 全局 token + 导航栏换肤（`app.wxss` brand token / `app.json` nav）— S ✅
- [ ] **T2** 自定义 TabBar + 残留硬编码 `#07c160` 清理 — M
- [x] **Checkpoint A**：✅ 测试全绿、业务代码无残留绿；⚠️ 配色观感待人工在 DevTools 确认

## Phase 1 — 首页启动页 + 菜单页拆分/改版（堂食路径端到端，高风险早做）
- [x] **T3** 菜单逻辑迁移到独立页 `pages/menu/index`（布局先不变，深链直达 menu）— L ✅
- [x] **T4** 首页重构为启动页（堂食/外卖两入口 + 切换；深链跳过启动页）— M ✅
- [x] **T5** 菜单页改版：左分类栏 + 右列表 + 顶部堂食/外卖切换（DESIGN.md 配色）— L ✅
- [ ] **Checkpoint B**：⏳ 自动化测试全绿（59）；⚠️ 待人工在微信开发者工具验证 启动页→堂食→扫码→新菜单→加购→订单确认→支付 全链路 + 深链回归

## Phase 2 — order_type 契约（后端字段 + 前端贯穿）
- [x] **T6** 后端 `Order.OrderType` + delivery 放开 `TableNo`（与 shansong 共享契约）— M ✅
- [x] **T7** 前端贯穿 order_type（首页/菜单→订单确认只读→createOrder）— M ✅
- [x] **Checkpoint C**：✅ 后端 build/test 全绿、堂食回归不变、堂食单携带 order_type

## Phase 3 — 真实 SKU / 规格（全栈，最大块）
- [x] **T8** 后端 `ProductSpec` 模型 + 下单按规格校验/计价 + OrderItem 记录规格 — L ✅
- [x] **T9** 管理后台规格管理（增删改 + api）— M ✅
- [x] **T10** 小程序「选规格」选择 + 购物车按 SKU 重构（含缓存迁移）— L ✅
- [x] **Checkpoint D**：✅ 后端单测覆盖含/无规格两路、前后端金额一致、cart_v2 版本化；⚠️ 端到端待 DevTools 验证

## Phase 4 — 外卖 地址→门店（衔接 shansong 规格）
- [x] **T11** Shop 地理字段 + 单门店解析（预留 nearest-shop 钩子）— M ✅
- [x] **T12** 外卖入口流程：chooseAddress → 门店 → 配送态菜单 → delivery 交接 — M ✅
- [x] **Checkpoint E**：✅ 自动化测试全绿、单门店解析、order_type=delivery 携带地址；⚠️ 真机 chooseAddress + 全链路待人工验证；运费/派单交 shansong

## 开工前待人工确认（见 plan.md Open Questions）
- [ ] Q1 order_type 事实来源 = 首页/菜单（订单确认只读）？
- [ ] Q2 `Order.OrderType`/`TableNo` 由本计划落地、shansong 消费？
- [ ] Q3 SKU 上线购物车缓存：清空 or 键版本化？
- [ ] Q4 无规格商品 = 单一默认规格/商品价？
- [ ] Q5 `custom-tab-bar/tab-bar/` 是否废弃、换肤跳过？
- [ ] Q6 外卖范围止于「地址→门店→delivery 交接」，运费/派单归 shansong？
