# 实施计划：商家后台「编辑菜品」内联规格(SKU)设置

## 目标
把现有**独立的「规格管理」弹层**合并进**新增/编辑菜品弹层**，让商家在一个界面里同时设置菜品信息 + 规格(SKU)。**复用已有接口，不改后端、不改小程序、不改规格数据模型。**

## 现状（已确认，勿改）
- 模型 `ProductSpec` = 扁平单维 `{id, name, price, status}`（1=上架/0=下架/2=售罄），`Product.Specs` 一对多。**本期不动。**
- 后端规格 CRUD 已就绪并复用：
  - `POST /merchant/products/:id/specs`（建规格，需已存在的 product id）
  - `PUT /merchant/specs/:id`、`DELETE /merchant/specs/:id`
  - api 封装：`admin/src/api/product.js` 的 `createProductSpec/updateProductSpec/deleteProductSpec`（**本期不改签名**）。
- 后台 `admin/src/views/Products.vue`：
  - 表格「规格(N)」列 → 点开**独立弹层** `specDialogVisible`，逐行即时调接口增改删。
  - 「编辑菜品」弹层 `dialogVisible`：名称/价格/分类/描述/图片/状态，保存走 `create/updateProduct`。
- 测试栈：Vitest + happy-dom，**无 @vue/test-utils**（不做组件挂载测试）。既有测试守护**纯函数/接口映射**（`utils/reward.test.js`、`api/product.test.js`、`api/client.test.js`）。

## 架构决策
1. **交互模型：本地草稿 + 保存时一次性落库（new 与 existing 统一）。**
   编辑弹层里规格改为本地可编辑列表（草稿），增删改不即时发请求；点「保存」时：
   先 `create/updateProduct` 拿到 product id → 再按草稿与原始规格的差异，调既有规格接口落库 → 刷新。
   - 这样**新菜品**也能在同一弹层里先填规格：保存时先建菜品、再用返回的 id 建规格（天然解决「建规格需 product id」）。
   - 取消则草稿丢弃，零副作用。
2. **差异计算抽成纯函数**（沿用小程序 `utils/spec.js` 的模式），用 Vitest 守护；组件只做薄装配（无组件测试栈）。
3. **移除独立规格弹层**，避免两处设置规格的入口重复/冲突；表格「规格」列降级为只读数量展示。
4. 失败处理：保存时规格接口按序 `await`，任一失败 → 提示 + `load()` 回读真实状态（管理后台可接受的最简策略；不引入事务/回滚）。

## 依赖图
```
T1 纯函数 diffSpecs/validateSpecs (admin/src/utils/specSync.js)
        │
        ▼
T2 编辑弹层内联「规格」草稿编辑区（展示+本地增删改，旧弹层暂留）
        │
        ▼
T3 保存时落库（菜品→规格 diff 落库，复用既有接口）+ 删除旧独立弹层 + 列降级
```

## 任务（垂直切片）

### T1 — 规格差异/校验纯函数 + 单测  〔S〕
**描述**：新增 `admin/src/utils/specSync.js`，把「草稿 vs 原始」的落库意图算成纯数据，供保存时驱动既有接口。
- `diffSpecs(original, draft)` → `{ creates:[{name,price,status}], updates:[{id,name,price,status}], deletes:[id] }`
  - draft 无 id 行 → creates；有 id 且 name/price/status 任一变化 → updates；original 有而 draft 缺的 id → deletes；未变化不产出。
- `validateSpecs(draft)` → `{ ok, message }`：每行 name 非空、price>0；空 draft 合法（= 无规格按菜品价售卖）。
**验收**：
- [ ] 新增行进 creates、改价进 updates、删除行进 deletes、未变不产出。
- [ ] 校验拦截空名/非正价，空列表判定合法。
**验证**：`cd admin && npm test`（新增 `specSync.test.js` 全绿，既有测试不回归）。
**依赖**：无。**文件**：`admin/src/utils/specSync.js`、`admin/src/utils/specSync.test.js`。

### T2 — 编辑弹层内联规格草稿编辑区  〔M〕
**描述**：在「新增/编辑菜品」弹层加「规格(SKU)」区：可编辑表（名称/价格/状态）+「添加一行」+「删除行」，全部操作**本地草稿** `form.specs`，暂不发请求。打开编辑时把 `row.specs` 深拷进草稿；新增时草稿为空。**旧独立弹层与「规格(N)」按钮本任务保持可用**（保证每步系统可用）。
**验收**：
- [ ] 打开「编辑」→ 规格区列出该菜品现有规格；改动只影响本地草稿，不发网络。
- [ ] 可加行/删行/改名改价改状态；空规格区显示「不设置则按菜品价售卖」。
- [ ] 「取消」后重开，草稿恢复为服务端数据（无脏留存）。
**验证**：`cd admin && npm run build` 通过；`npm run dev` 手测：编辑弹层规格区增删改纯本地，控制台无报错，Network 无规格请求。
**依赖**：T1（保存阶段用，本任务先备好草稿结构）。**文件**：`admin/src/views/Products.vue`。

### T3 — 保存落库 + 收口  〔M〕
**描述**：`save()` 改为：① `create/updateProduct`（新建取返回 id）；② `validateSpecs` 不过则中止提示；③ `diffSpecs(original, draft)` → 依次 `createProductSpec(pid,...)`/`updateProductSpec(id,...)`/`deleteProductSpec(id)`；④ `load()` 刷新并关弹层。随后**删除独立规格弹层**相关状态/模板/方法（`specDialogVisible`、`openSpecs`、`addSpec`、`saveSpec`、`removeSpec`、`reloadSpecProduct`、`newSpec`），表格「规格」列改为只读 `规格({{row.specs?.length||0}})`。
**验收**：
- [ ] 新建菜品 + 2 条规格 → 落库后列表显示该菜品且规格数=2，小程序读到一致价格。
- [ ] 编辑：改一条规格价、删一条、加一条 → 保存后服务端与界面一致。
- [ ] 规格接口失败 → 提示且 `load()` 回读真实状态，不留半保存的界面假象。
- [ ] 旧独立规格弹层入口与代码已移除；`npm run build` 无未用变量/导入告警。
**验证**：`cd admin && npm test`（T1 守护不回归）+ `npm run build`；`npm run dev` 走查新建/编辑两条链路（含一次故意失败：如把某规格价改成 0 被校验拦下）。
**依赖**：T1、T2。**文件**：`admin/src/views/Products.vue`。

### Checkpoint A（收尾）
- [ ] `cd admin && npm test` 全绿、`npm run build` 干净（无新告警）。
- [ ] 手测四条：新建带规格 / 编辑改规格 / 删规格 / 取消零副作用。
- [ ] 后端、小程序、规格模型与接口签名**零改动**（git diff 仅含 `admin/`）。
- [ ] 人工 review 后再决定是否构建上传后台（部署见 deploy 记忆）。

## 守护测试映射
- T1 → `admin/src/utils/specSync.test.js`（纯函数）。
- T2/T3 → 无组件测试栈，靠 `npm run build` + 手测；落库意图由 T1 的 diff 守护，接口映射由既有 `api/product.test.js` 守护。

## 不在本期
- 后端模型/接口改动、库存(stock)、多属性 SKU 组合矩阵（杯型×温度×糖度）、小程序改动。
- 引入 @vue/test-utils 做组件测试（如需再单列）。

## 风险
| 风险 | 级别 | 缓解 |
|---|---|---|
| 保存中途规格接口失败留下半保存状态 | 中 | 按序 await + 失败即 `load()` 回读；管理后台可重试 |
| 无组件测试，UI 回归靠手测 | 中 | 把落库意图收敛到 T1 纯函数守护；保存路径薄装配 |
| 移除旧弹层牵连未清干净的状态/方法 | 低 | T3 验收含 build 无未用告警 |

## 开放问题
- 失败策略是否需要「全有或全无」（前端模拟事务）？默认按序+回读，足够管理后台场景；如需更强一致再提。
