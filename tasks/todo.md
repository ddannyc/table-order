# Todo: 去定制化 — 全面改用 weui 原生组件 + 默认配色

详见 `tasks/plan.md`。基准色 = weui 默认绿 `#07c160`（仅浅色）；菜单重构为纯 weui（顶部分类页签 + 媒体列表）；TabBar 换 weui tabbar。token 删除放最后。

## Phase 0 — weui 组件注册（加法）
- [x] **T1** app.json 注册所需 weui 组件（tabbar/searchbar/half-screen-dialog/dialog/grids…）— S ✅
- [x] **Checkpoint A**：✅ 编译/测试通过、页面不变

## Phase 1 — 底部导航换 weui tabbar
- [x] **T2** weui tabbar 替换 custom-tab-bar（home/invite/menu/profile）— M ✅
- [x] **Checkpoint B**：✅ 四页 weui tabbar、82 测试全绿；⚠️ 待人工 DevTools 确认图标/高度 —— **在此暂停等方向确认**

## Phase 2 — 逐页迁移到 weui + 默认色（每页一纵切）
- [ ] **T3** login → weui — S
- [ ] **T4** home 启动页（两入口 grids/cells）→ weui — M
- [ ] **T5** menu 重构为纯 weui（navbar 页签 + 媒体列表 + half-screen-dialog 选规格 + 结算条）+ 重写 menu 测试 — L
- [ ] **T6** order-confirm → weui — M
- [ ] **T7** profile → weui — M
- [ ] **T8** invite → weui — M
- [ ] **T9** share-code → weui — S
- [ ] **Checkpoint C**：全页 weui、`grep var(--brand` 无命中、jest 全绿

## Phase 3 — 移除定制体系
- [ ] **T10** 删 DESIGN.md + app.wxss token + design-tokens 测试 + 废弃 custom-tab-bar；清理 DESIGN.md 引用 — M
- [ ] **Checkpoint D**：无 DESIGN/无 token/无 custom-tab-bar、全站 weui 默认绿、jest 全绿、人工验收

## 待人工确认（见 plan.md Open Questions）
- [ ] Q1 菜单顶部是否要 weui-search-bar？
- [ ] Q2 profile/invite 渐变头部接受改 weui 纯色简化？
- [ ] Q3 仅浅色、不预留深色钩子？
