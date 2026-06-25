# Todo: 配色改版 — 暖橙方案（取自 ui-add-food.png）

详见 `tasks/plan.md`。纯换色，不动布局/交互。DESIGN.md 仍是唯一真源。
品牌橙 `#F88818`；金黄→桃橙渐变；深墨文字 `#083038`。

## Phase 1 — 暖橙配色改版
- [ ] **T1** 重写 DESIGN.md 配色为暖橙（含 `--brand-gradient`，来源改 ui-add-food.png）— S
- [ ] **T2** 应用 `app.wxss` token 值 + 5 处导航栏 hex + TabBar 兜底 + design-tokens 测试（含无残留 `#189CA8` 扫描）— M
- [ ] **T3** 启动页 hero / 菜单门店栏用暖色 banner 渐变贴合图片顶部 — S
- [ ] **Checkpoint**：jest 全绿、无残留 `#189CA8`、⚠️ 人工 DevTools 确认观感「好看」且可读

## 待人工确认（见 plan.md Open Questions）
- [ ] Q1 价格用深墨（忠于图片）还是改回橙色？
- [ ] Q2 品牌橙 `#F88818` 是否需更柔和/更深？
- [ ] Q3 成功色改绿 `#2BA471` 是否可接受？
