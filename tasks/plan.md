# 实施计划：优化「外卖」入口的白屏/加载体验

## 问题诊断（事实来源）
点「外卖」比点「堂食」明显多一段「白屏加载中」，根因是外卖比堂食多了**一整段导航前的静默网络请求**：

- **堂食** `scanDineIn()`：扫码 → **立即** `wx.reLaunch` 进菜单 → 菜单显示「加载中…」转圈 → 拉数据。一段等待。
- **外卖** `chooseDelivery()`：点按 → 先发 `resolveDeliveryShop()`（`GET /delivery/shop`），**这期间无任何反馈**（人停在首页）→ 回来后才 `wx.reLaunch`（重建页面闪一帧）→ 菜单再 `Promise.all([getShop, getShopProducts])` → 「加载中…」。两段等待，第一段零反馈。

附带浪费：`ResolveDeliveryShop` 后端返回的就是 `toPublicShopDTO(shop)`，与 `GetShop` **完全同一个 DTO**；菜单又 `getShop(shopId)` 拉了一遍 —— 外卖路径实际发了 **3 个请求**（`/delivery/shop` → `getShop` ∥ `getShopProducts`），其中 `getShop` 冗余。

**已排除项**：`app.json` 的 `window.backgroundColor` 已是 `#FBEFF3`，菜单页未覆盖、继承生效，故 reLaunch 那一帧本就是品牌浅粉而非纯白 → **无需改 backgroundColor**（原 Fix 3 作废）。菜单初始 `data.loading:true` 且首屏即命中 `wx:elif="{{loading}}"` 转圈分支，无空白内容帧。

## 目标
让「外卖」与「堂食」一样**点按即跳转**：去掉导航前的静默请求，把门店解析挪进菜单页，由菜单已有的「加载中…」转圈盖住整段等待；并复用 `/delivery/shop` 已返回的门店 DTO，跳过冗余的 `getShop`，把外卖路径从 3 个请求降到 2 个。

范围：**仅 `frontend/`**，不改后端、不改接口签名、不改数据模型。

## 架构决策
1. **解析下沉到菜单页**：菜单 `onLoad` 在 `order_type=delivery` 且**无 `shop_id`** 时，自己 `resolveDeliveryShop()` → `loadData(shop)`。首页点按只负责立即 `reLaunch`，无网络。
2. **DTO 透传跳过冗余请求**：`loadData(prefetchedShop?)` —— 传入门店则 `Promise.resolve(shop)`，不再 `getShop`；不传（堂食/重试）维持原行为。`/delivery/shop` 与 `getShop` 同为 `toPublicShopDTO`，字段一致，安全。
3. **菜单兼容两条外卖入口**：保留既有 `delivery && shop_id` 分支（向后兼容/可能的深链），新增 `delivery && !shop_id` 解析分支。T1 加完后系统对新旧 URL 都可用。
4. **错误与重试归位菜单**：无可配送门店时，由菜单 `setData({loading:false, error:true})` + toast「暂无可配送门店」，`onRetry` 在「外卖且未绑定门店」时重走解析。无门店提示从首页迁到菜单。
5. **测试策略**：小程序页面无组件挂载测试栈；沿用仓库既有 jest 模式（mock `wx`/api 模块后直接 `pageConfig.method.call(ctx, ...)` 断言）。`home-launcher.test.js` / `menu-page.test.js` 已覆盖这两页的导航装配，按 TDD 先改测试（RED）再改实现（GREEN）。

## 依赖图
```
T1 菜单页承接外卖冷启动（解析门店 + DTO 透传跳过 getShop + 错误/重试）  [additive, 向后兼容]
        │
        ▼
T2 首页「外卖」即时跳转（去掉导航前静默请求；无门店提示迁至菜单）       [行为切换]
```

## 任务（垂直切片）

### T1 — 菜单页承接外卖冷启动  〔M〕
**描述**：菜单页新增「无 shop_id 的外卖冷启动」能力，并复用 `/delivery/shop` 的门店 DTO 跳过冗余 `getShop`。保留既有 `delivery && shop_id` 分支不动（向后兼容）。
- `menu/index.js`：`require` 增加 `resolveDeliveryShop`（来自 `../../api/index.js`）。
- `onLoad`：新增分支 —— `options.order_type === 'delivery' && !options.shop_id` → `this.setData({ orderType:'delivery' })` 后 `this.loadDeliveryShop()`。
- 新增 `loadDeliveryShop()`：`setData({loading:true, error:false})` → `resolveDeliveryShop()`：成功 `setData({ boundShopId: shop.id, boundTableNo:'', orderType:'delivery' })` 并 `this.loadData(shop)`；失败 `setData({loading:false, error:true})` + `wx.showToast({title:'暂无可配送门店', icon:'none'})`。
- `loadData(prefetchedShop)`：`const shopP = prefetchedShop ? Promise.resolve(prefetchedShop) : getShop(boundShopId)`；其余不变。
- `onRetry()`：`if (this.data.orderType === 'delivery' && !this.data.boundShopId) this.loadDeliveryShop(); else this.loadData()`。

**验收**：
- [ ] `onLoad({order_type:'delivery'})`（无 shop_id）→ 调用 `loadDeliveryShop`，不直接 `loadData`。
- [ ] `loadDeliveryShop` 成功 → `loadData` 收到门店对象，且**不**调用 `getShop`（只 `getShopProducts`）。
- [ ] `loadDeliveryShop` 失败 → `loading:false, error:true` 且弹出「暂无可配送门店」。
- [ ] 既有 `onLoad({order_type:'delivery', shop_id:'7'})` 行为不变（`menu-page.test.js` 不回归）。
- [ ] `onRetry` 在外卖未绑定门店时走 `loadDeliveryShop`，否则走 `loadData`。

**验证**：`cd frontend && npm test`（新增/扩展菜单测试全绿，既有不回归）。无 build 步骤。
**依赖**：无。**文件**：`frontend/pages/menu/index.js`、`frontend/__tests__/menu-page.test.js`（或新增 `frontend/__tests__/delivery-cold-start.test.js`）。

### T2 — 首页「外卖」即时跳转  〔S〕
**描述**：首页点「外卖」改为**立即** `reLaunch` 进菜单的外卖模式，不再导航前解析门店；无门店提示由菜单承担（T1 已实现）。
- `home/index.js`：`chooseDelivery()` → 直接 `wx.reLaunch({ url: '/pages/menu/index?order_type=delivery' })`。
- 移除 `home/index.js` 中现已未用的 `resolveDeliveryShop` 导入。

**验收**：
- [ ] 点「外卖」→ 同步 `reLaunch` 到 `/pages/menu/index?order_type=delivery`，URL 不含 `shop_id`。
- [ ] `chooseDelivery` 不再调用 `resolveDeliveryShop`（导航前零网络）。
- [ ] 堂食路径（`scanDineIn`/`tabChange`）不受影响。

**验证**：`cd frontend && npm test`（先改 `home-launcher.test.js` 为新行为=RED，再改实现=GREEN；删除「首页无门店提示」用例，其意图已由 T1 菜单测试覆盖）。
**依赖**：T1（菜单须先能处理无 shop_id 的外卖入口）。**文件**：`frontend/pages/home/index.js`、`frontend/__tests__/home-launcher.test.js`。

### Checkpoint A（收尾）
- [ ] `cd frontend && npm test` 全绿（含改写后的 home/menu 用例）。
- [ ] 手测三条（DevTools/真机）：① 点外卖 → 立即转圈 → 出菜单（无静默卡顿）；② 后端无可配送门店 → 菜单 error 态 + toast，点重试可重走解析；③ 堂食扫码/深链不受影响。
- [ ] 后端、接口签名、数据模型**零改动**（git diff 仅含 `frontend/`）。
- [ ] 人工 review 后再决定是否上传小程序（部署见 deploy 记忆；**停等用户**）。

## 守护测试映射
- T1 → `menu-page.test.js` / `delivery-cold-start.test.js`（onLoad 分支、loadDeliveryShop 成功/失败、loadData 跳过 getShop、onRetry 分流）。
- T2 → `home-launcher.test.js`（chooseDelivery 即时导航、不再解析门店）。
- 既有 `api-delivery.test.js` / `order-confirm-delivery.test.js` 守护下游下单逻辑，本期不应触及。

## 不在本期
- 后端合并端点（一次返回门店+菜单）、菜单骨架屏（skeleton）替换转圈、`reLaunch`→`navigateTo` 转场动画改造、预拉取/缓存门店。
- 任何后端/模型/接口签名改动。

## 风险
| 风险 | 级别 | 缓解 |
|---|---|---|
| `/delivery/shop` 与 `getShop` DTO 未来分叉，透传漏字段 | 低 | 两者现同为 `toPublicShopDTO`；变更需同步，已在文档标注 |
| 改写 `home-launcher.test.js` 误删有效断言 | 低 | 无门店用例的意图迁到 T1 菜单测试，不丢覆盖 |
| 菜单 `delivery && shop_id` 旧分支变僵尸 | 低 | 保留作向后兼容；如确认无来源再单独清理 |

## 开放问题
- 是否要顺带把菜单「加载中…」转圈换成骨架屏以进一步降低「空等」感？默认不做（超范围），如要再单列。
