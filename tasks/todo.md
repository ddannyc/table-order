# Todo：菜单左右布局 + 入口即定单型（含前端设计规范）

详见 `tasks/plan.md`。一任务一提交，TDD（RED→GREEN→回归 jest）。
基线：绝不 git add `frontend/config.js` / `.claude/` / `specs/`。

## 任务（纵切，按依赖顺序）
- [x] **M1** 移除菜单切换 + 只读模式标识 + 单型可靠推导 — M ✅
- [x] **M2** 购物车按 shopId+orderType 隔离（storage/product/menu/order-confirm）— M（依赖 M1）✅
- [ ] **M3** CSS/SVG 分类占位图 + 图片兜底助手（无二进制资源）— S（独立）
- [ ] **M4** 左右布局 + 大图卡片（左类目轨 / 右卡片，激活态绿竖条）— L（依赖 M1、M3）

## Checkpoint
- [ ] **M5** 设计自审（/frontend-design）+ 全量 jest 绿（≥95）+ 模拟器截图核对

## 不在本期
- 真实菜品摄影 / 后台图片上传
- 类目 scroll-spy / 锚点联动
- 菜单内"切换点餐方式"入口（首页唯一入口）
