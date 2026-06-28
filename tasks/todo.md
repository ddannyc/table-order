# Todo：商家后台「编辑菜品」内联规格(SKU)设置

详见 `tasks/plan.md`。范围：**仅 `admin/`**，复用既有规格接口，不改后端/小程序/规格模型。
TDD（纯函数先行）、一任务一提交；改代码同提交内同步守护测试。
**git 卫生**：只 stage 该任务文件；禁止 `git add -A`；绝不 stage `frontend/config.js`/`.claude/`。
分支：建议 `feat/admin-inline-spec`（开工前切）。

## 任务
- [x] **T1** 规格差异/校验纯函数 `admin/src/utils/specSync.js`（`diffSpecs`/`validateSpecs`）+ `specSync.test.js` — S ✅
- [x] **T2** 编辑弹层内联「规格」草稿编辑区（展示+本地增删改，旧弹层暂留）— M ✅
- [x] **T3** 保存落库（菜品→规格 diff 复用既有接口）+ 删除旧独立规格弹层 + 「规格」列降级只读 — M ✅
- [~] **Checkpoint A** `npm test` 全绿 ✅ + `npm run build` 干净 ✅ + git diff 仅含 `admin/`+`tasks/` ✅ + 手测四条链路 ⏳(需起 admin dev + 后端登录，留人工) + 部署 ⏳(停等用户)

## 守护测试映射
- T1 → `admin/src/utils/specSync.test.js`
- T2/T3 → 无组件测试栈：`npm run build` + 手测；落库意图由 T1 守护，接口映射由既有 `api/product.test.js` 守护

## 不在本期
- 后端/模型/接口改动、库存、多属性 SKU 组合矩阵、小程序改动、引入 @vue/test-utils
