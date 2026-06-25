# Implementation Plan: 堂食/外卖双模式 + 菜品页改版 + SKU + 全站换肤

## Overview
将现有「扫码堂食单菜单」小程序升级为堂食/外卖双模式点餐应用：入口式首页（堂食/外卖两入口 + 切换）、独立的左分类栏+右列表菜品页（参考 `ui-add-food.png`）、真实规格(SKU)选择，并将全站从微信绿换肤为 `DESIGN.md` 的 teal 主色 + 橙强调。前端为微信原生小程序，后端 Go(Gin+GORM+PostgreSQL)，管理后台 Vue。

设计规范见根目录 `DESIGN.md`（颜色取自 `ui-color.png`，唯一真源）。本计划严格分阶段、纵向切片，每阶段后留检查点；除非检查点通过，不进入下一阶段。

## Architecture Decisions
- **order_type 事实来源 = 首页/菜单的选择**（推荐）。订单确认页对下单类型只读反映，不再作为切换点。理由：避免首页/菜单与订单确认页出现互相矛盾的 `order_type` 状态。**待人工确认**（Open Q1）。
- **`Order.OrderType` + `TableNo` 放开为本计划与 `specs/shansong-delivery.md` 的共享契约**。在 Phase 2 一次性定义；闪送外卖配送（运费/派单/回调）仍归属 shansong 规格，本计划止于「地址→门店」交接。
- **菜单从首页拆出为独立页 `pages/menu/index`**。首页改为入口启动页。扫码深链（`shop_id`+`table_no`）必须自动跳过启动页直达堂食菜单，避免回归。
- **SKU 采用 `ProductSpec` 子表**（1 商品 N 规格）。无规格的旧商品视为「单一默认规格」，保持兼容。购物车与订单行项的键从 `product_id` 迁移为 `product_id + spec_id`。
- **「外卖=选地址定位门店」当前落到唯一门店**（单门店现实）。`Shop` 增加经纬度字段并预留 nearest-shop 钩子，但不做多门店排序（见 Not Doing）。
- **换肤为机械化 token 替换**：组件只引用 `app.wxss` 的 CSS 变量，不写死色值。

## Dependency Graph
```
DESIGN.md (已完成)
   │
   └── Phase 0 换肤 (app.wxss tokens / app.json nav / tab-bar / 残留硬编码色)   [横切地基, 低风险]
          │
          ├── Phase 1 首页启动页 + 菜单页拆分/改版 (堂食路径端到端)            [结构性, 高风险, 早做]
          │       │
          │       └── Phase 2 order_type 契约 (后端字段 + 前端贯穿)            [薄纵切, 解锁外卖/反映选择]
          │               │
          │               ├── Phase 3 真实 SKU/规格 (后端模型 + 后台 + 前端选规格)  [全栈, 最大]
          │               │
          │               └── Phase 4 外卖 地址→门店 (Shop 地理 + chooseAddress)    [依赖 Shop 地理; 衔接 shansong]
```
实现顺序自底向上：先地基(换肤)，再结构(首页/菜单)，再契约(order_type)，再并行的两条增量(SKU / 外卖)。

---

## Task List

### Phase 0: 品牌换肤地基（横切，低风险）

#### Task 1: 全局 token 与导航栏换肤
**Description:** 按 `DESIGN.md` 把 `app.wxss` 的 `page{}` 设计 token 从微信绿替换为 teal/橙体系，新增品牌 token（`--brand-primary` 等），并更新 `app.json` 的 `navigationBarBackgroundColor`。
**Acceptance criteria:**
- [ ] `app.wxss` 中 `--weui-primary` 及相关变量改为 `#189CA8` 体系，新增 DESIGN.md 列出的 `--brand-*` token。
- [ ] `app.json` `navigationBarBackgroundColor` = `#189CA8`。
- [ ] 不再有页面依赖旧绿色 token 值（变量名保留，值更新）。
**Verification:**
- [ ] 微信开发者工具编译无报错。
- [ ] 人工：首页/我的/邀请导航栏与主按钮显示为 teal。
**Dependencies:** None
**Files likely touched:** `frontend/app.wxss`, `frontend/app.json`
**Estimated scope:** S

#### Task 2: 自定义 TabBar 与残留硬编码色换肤
**Description:** 替换 `custom-tab-bar-comp` 选中态 `#07c160`，并清理 `invite/profile/share-code/order-confirm/home` 等页面 wxss/json 中残留的硬编码 `#07c160`，统一引用 token。
**Acceptance criteria:**
- [ ] `grep 07c160` 在 `frontend/`（排除 `miniprogram_npm/weui-miniprogram/` 库与已确认的废弃 `custom-tab-bar/tab-bar/`，见 Open Q5）无业务代码命中。
- [ ] TabBar 选中态为 `--brand-primary`。
**Verification:**
- [ ] 编译无报错。
- [ ] 人工：切换三个 Tab，选中色为 teal；各页面无残留绿色。
**Dependencies:** Task 1
**Files likely touched:** `frontend/miniprogram_npm/custom-tab-bar-comp/index.wxss`, `frontend/pages/{invite,profile,share-code,order-confirm,home}/index.{wxss,json}`
**Estimated scope:** M

#### Checkpoint A — 换肤
- [ ] 编译干净，所有页面渲染正常。
- [ ] 全站无残留微信绿（业务代码层面）。
- [ ] 与人工确认配色观感符合 `DESIGN.md` 后再进入 Phase 1。

---

### Phase 1: 首页启动页 + 菜单页拆分与改版（堂食路径端到端）

#### Task 3: 菜单逻辑迁移到独立页 `pages/menu/index`
**Description:** 把当前 `pages/home/index` 的「已绑桌→点餐」逻辑（菜品加载、加购、购物车条、去结算）整体迁移到新建 `pages/menu/index`，**布局先保持不变**。调整路由：扫码深链与桌号绑定后进入 menu。注册到 `app.json` pages。
**Acceptance criteria:**
- [ ] 新页面 `pages/menu/index` 复刻原点餐功能，加购/数量联动/去结算正常。
- [ ] 扫码或带 `shop_id`+`table_no` 深链进入小程序时，最终停留在 menu 且菜品正确加载。
- [ ] `order-confirm` 的返回/跳转路径同步更新（原指向 home 的改为 menu，如适用）。
**Verification:**
- [ ] 人工：扫码 → 菜单 → 加购 → 去结算 → 订单确认 全链路通。
- [ ] 编译无报错。
**Dependencies:** Checkpoint A
**Files likely touched:** `frontend/pages/menu/index.*`（新）, `frontend/app.json`, `frontend/pages/order-confirm/index.js`
**Estimated scope:** L（如超范围，按「迁移」与「路由」拆两子任务）

#### Task 4: 首页重构为启动页（堂食/外卖两入口 + 切换）
**Description:** 将 `pages/home/index` 改为入口启动页：两张入口卡片（堂食/外卖）+ 分段切换控件（按 `DESIGN.md`）。堂食卡 → 触发扫码 → menu。外卖卡 → 占位（Phase 4 接入，先禁用或提示「敬请期待」）。深链命中时自动跳过启动页直达 menu(堂食)。
**Acceptance criteria:**
- [ ] 首页展示堂食/外卖两入口，符合 DESIGN.md 视觉。
- [ ] 点堂食 → 扫码绑定 → 进入 menu。
- [ ] 带 `shop_id`+`table_no` 启动时不停留在启动页，直达 menu。
- [ ] 外卖入口为占位态（不报错、不进入未完成流程）。
**Verification:**
- [ ] 人工：冷启动看到启动页；扫码深链直达菜单；堂食路径通。
**Dependencies:** Task 3
**Files likely touched:** `frontend/pages/home/index.*`
**Estimated scope:** M

#### Task 5: 菜单页改版为左分类栏 + 右列表 + 顶部堂食/外卖切换
**Description:** 按 `ui-add-food.png` 重做 `pages/menu/index` 布局：左侧分类竖栏（选中态主色竖条）、右侧菜品列表（图/名/规格描述/橙色价格/加购）、顶部门店信息条 + 堂食/外卖胶囊切换；全部按 `DESIGN.md` 配色。**仍用现有 +/- 加购**（SKU 在 Phase 3）。
**Acceptance criteria:**
- [ ] 左栏分类点击联动右侧列表滚动/筛选。
- [ ] 顶部堂食/外卖切换可改变当前 `order_type`（UI 状态；后端贯穿在 Phase 2）。
- [ ] 视觉符合 DESIGN.md（teal 主色、橙价格/徽标）。
**Verification:**
- [ ] 人工：分类切换、加购、购物车条、去结算均正常；多分类滚动正确。
**Dependencies:** Task 4
**Files likely touched:** `frontend/pages/menu/index.{wxml,wxss,js}`
**Estimated scope:** L

#### Checkpoint B — 堂食新 UI 端到端
- [ ] 启动页 → 堂食 → 扫码 → 新版菜单 → 加购 → 订单确认 → 支付 全链路通。
- [ ] 扫码深链回归正常。
- [ ] 与人工确认菜单页观感后进入 Phase 2。

---

### Phase 2: order_type 契约（后端字段 + 前端贯穿）

#### Task 6: 后端新增 `Order.OrderType` 并放开 delivery 的 TableNo
**Description:** `models/order.go` 新增 `OrderType`（`dine_in`|`delivery`，默认 `dine_in`），注册迁移。`CreateOrderRequest` 接收 `order_type`；当 `delivery` 时 `TableNo` 允许为空（堂食仍必填）。**与 `specs/shansong-delivery.md` 共享此契约**。
**Acceptance criteria:**
- [ ] `Order` 含 `OrderType`，迁移生效，旧数据默认 `dine_in`。
- [ ] `dine_in` 下单行为与现状完全一致（金额/返利/福利金不变）。
- [ ] `delivery` 下单允许空 `table_no`（本阶段仅契约，前端 Phase 4 才真正发 delivery）。
**Verification:**
- [ ] `cd backend && go build ./... && go test ./...` 通过。
- [ ] 新增/更新单测覆盖：dine_in 回归、delivery 空 table_no 校验。
**Dependencies:** Checkpoint B（亦可与 Phase 1 并行，但需先定契约）
**Files likely touched:** `backend/models/order.go`, `backend/api/handler/order.go`, `backend/config/database.go`, `backend/api/handler/*_test.go`
**Estimated scope:** M
**Coordination:** 与 shansong 规格作者确认由本任务落地该字段，shansong 仅消费（避免重复迁移）。

#### Task 7: 前端贯穿 order_type（首页/菜单 → 订单确认 → createOrder）
**Description:** 将 `order_type` 从首页/菜单选择贯穿到 `order-confirm` 与 `api/index.js` `createOrder`。订单确认页对下单类型**只读展示**（事实来源在首页/菜单）。
**Acceptance criteria:**
- [ ] `createOrder` 入参新增 `order_type`，签名向后兼容（默认 dine_in）。
- [ ] 堂食单提交 `order_type=dine_in`，行为不变。
- [ ] 订单确认页显示当前下单类型，不可在此切换。
**Verification:**
- [ ] 人工：堂食下单，抓包/日志确认 `order_type=dine_in`。
**Dependencies:** Task 6
**Files likely touched:** `frontend/api/index.js`, `frontend/pages/order-confirm/index.*`, `frontend/pages/menu/index.js`
**Estimated scope:** M

#### Checkpoint C — 契约
- [ ] 后端 build/test 全绿；堂食回归不变。
- [ ] 堂食单正确携带 `order_type`。

---

### Phase 3: 真实 SKU / 规格（全栈，最大块）

#### Task 8: 后端 ProductSpec 模型与下单按规格校验
**Description:** 新增 `models/product_spec.go`（`ID, ProductID, Name, Price, Status`）+ 迁移；`GetShopProducts` 返回每商品的 specs；`CreateOrderRequest.Items` 支持 `spec_id`，按规格价校验与计价；`OrderItem` 新增 `SpecID`/`SpecName`。无规格商品按「单一默认规格」处理（用商品自身价）。
**Acceptance criteria:**
- [ ] 新增 `product_specs` 表，迁移生效。
- [ ] 含规格商品下单按所选规格价计；无规格商品下单仍按商品价（回归）。
- [ ] `OrderItem` 记录规格名/价。
**Verification:**
- [ ] `go build ./... && go test ./...` 通过；新增单测覆盖含规格/无规格两路。
**Dependencies:** Checkpoint C
**Files likely touched:** `backend/models/product_spec.go`(新), `backend/models/order_item.go`, `backend/api/handler/{product,order}.go`, `backend/config/database.go`, `*_test.go`
**Estimated scope:** L

#### Task 9: 管理后台规格管理
**Description:** `admin` 商品管理支持为商品增删改规格（名称+价格+状态），对接后端规格接口。
**Acceptance criteria:**
- [ ] 商家可为商品添加/编辑/删除规格并持久化。
- [ ] 无规格商品仍可正常创建/编辑（向后兼容）。
**Verification:**
- [ ] 人工：后台添加规格 → 小程序菜单可见该规格。
- [ ] `cd admin && npm run build` 通过。
**Dependencies:** Task 8
**Files likely touched:** `admin/src/views/Products.vue`, `admin/src/api/product.js`
**Estimated scope:** M

#### Task 10: 小程序「选规格」选择与购物车按 SKU 重构
**Description:** 菜单页为含规格商品显示「选规格」按钮 → 弹出规格选择层；购物车键从 `product_id` 迁移为 `product_id+spec_id`（`utils/storage.js` + `api/product.js` + menu + order-confirm），下单携带 `spec_id`。处理旧购物车缓存迁移（升级时清空或键版本化）。
**Acceptance criteria:**
- [ ] 含规格商品需选规格后加购；不同规格在购物车独立成行。
- [ ] 无规格商品仍可直接 +/- 加购。
- [ ] 订单确认与下单按所选规格价计算，与后端一致。
**Verification:**
- [ ] 人工：选规格 → 加购 → 不同规格独立 → 下单金额正确 → 订单详情含规格。
**Dependencies:** Task 9
**Files likely touched:** `frontend/utils/storage.js`, `frontend/api/product.js`, `frontend/pages/menu/index.*`, `frontend/pages/order-confirm/index.*`, `frontend/api/index.js`
**Estimated scope:** L

#### Checkpoint D — SKU
- [ ] 含规格与无规格商品两条路径均端到端通过，金额前后端一致。
- [ ] 旧缓存购物车不导致崩溃。

---

### Phase 4: 外卖 地址→门店（衔接 shansong 规格）

#### Task 11: Shop 地理字段 + 单门店解析
**Description:** `models/shop.go` 新增经纬度/配送相关字段 + 迁移 + 后台维护；提供门店解析（当前单门店直接返回，预留 nearest-shop 钩子）。
**Acceptance criteria:**
- [ ] `Shop` 含 lat/lng（及必要配送字段），迁移生效，后台可维护。
- [ ] 提供「按地址解析门店」入口，当前返回唯一门店。
**Verification:**
- [ ] `go build ./... && go test ./...` 通过。
**Dependencies:** Checkpoint D（亦与 shansong 规格 Open Q1 对齐）
**Files likely touched:** `backend/models/shop.go`, `backend/api/handler/shop.go`, `backend/config/database.go`, `admin/src/views/ShopSettings.vue`
**Estimated scope:** M
**Coordination:** 与 `specs/shansong-delivery.md` Open Q1（寄件门店坐标）合并实现。

#### Task 12: 外卖入口流程（chooseAddress → 门店 → 配送态菜单）
**Description:** 首页外卖入口：`wx.chooseAddress` 选地址（缓存最近一次作默认）→ 解析门店 → 进入 menu(配送态，无桌号)。`app.json` 声明 `requiredPrivateInfos:["chooseAddress"]`。在 order-confirm 把地址 + `order_type=delivery` 交接给 shansong 流程（**不含运费/派单**）。
**Acceptance criteria:**
- [ ] 外卖入口可选地址并缓存默认；进入菜单为配送态（无桌号）。
- [ ] 提交订单 `order_type=delivery` 且携带地址；运费/派单留给 shansong 规格。
**Verification:**
- [ ] 真机/体验版：chooseAddress 授权回填 → 菜单 → 订单确认携带地址与 delivery 类型。
**Dependencies:** Task 11
**Files likely touched:** `frontend/pages/home/index.*`, `frontend/pages/menu/index.js`, `frontend/pages/order-confirm/index.*`, `frontend/app.json`, `frontend/utils/storage.js`
**Estimated scope:** M

#### Checkpoint E — 外卖交接
- [ ] 外卖路径到达订单确认页，携带地址与 `order_type=delivery`。
- [ ] 堂食路径回归全绿。
- [ ] 交接点与 `specs/shansong-delivery.md` 对齐，后续运费/派单按该规格推进。

---

## Risks and Mitigations
| Risk | Impact | Mitigation |
|------|--------|------------|
| 首页=菜单拆分破坏扫码深链 | High | Task 3/4 明确深链自动跳过启动页直达 menu；Checkpoint B 专项回归 |
| 与 shansong 规格在 `Order.OrderType`/`TableNo` 重复迁移 | High | Phase 2 一次性定义共享契约，shansong 仅消费；任务内标注 Coordination |
| 购物车键 `product_id→sku` 破坏存量缓存 | Med | Task 10 升级时清空或键版本化；Checkpoint D 验证旧缓存不崩 |
| 全站换肤回归面广 | Med | Phase 0 先行、独立、低风险；逐页人工核验 |
| 单门店下「最近门店」无意义 | Med | 仅做单门店解析 + 钩子，不做排序（Not Doing） |
| chooseAddress 需隐私配置 + 真机验证 | Med | Task 12 同步 app.json + 小程序后台隐私指引；真机验收 |
| 后端金额以服务端为准，规格价需双端一致 | Med | Task 8 服务端按 spec 重算金额，单测覆盖；前端仅展示 |

## Open Questions
1. **order_type 事实来源**：确认以「首页/菜单选择」为准、订单确认页只读？（计划默认如此）
2. **共享契约归属**：`Order.OrderType` + `TableNo` 放开由本计划 Phase 2 落地、shansong 规格消费 —— 是否同意此分工与合并顺序？
3. **购物车缓存迁移**：SKU 上线时存量购物车缓存采用「清空」还是「键版本化」？
4. **无规格商品语义**：确认无规格商品按「单一默认规格 / 商品自身价」处理（保持兼容）？
5. **废弃 TabBar**：`frontend/miniprogram_npm/custom-tab-bar/tab-bar/` 是否为废弃代码（`app.json` 仅用 `custom-tab-bar-comp`）？换肤是否跳过它？
6. **外卖范围与 shansong 衔接时点**：Phase 4 仅做到「地址→门店→delivery 交接」，运费/派单完全交给 shansong 规格 —— 范围是否如此？

## Not Doing（及原因）
- **多门店「最近门店」排序** —— 当前单门店，仅留钩子，不做排序。过早。
- **后端用户地址簿** —— shansong 规格已定仅本地缓存。
- **配送费 / 闪送派单 / 回调** —— 归属 `specs/shansong-delivery.md`，本计划止于交接。
- **管理后台(Vue) 视觉换肤** —— 需求针对小程序；后台 UI 不在范围（仅为 SKU/门店地理新增表单字段）。
