# Todo：复刻 mp-ui（首页 / 菜品 / 下单 / Tab）

详见 `tasks/plan.md`。来源 spec：`docs/ideas/replicate-mp-ui.md`。
一任务一提交，TDD（RED→GREEN→回归 jest）。
基线：绝不 git add `frontend/config.js` / `.claude/` / `specs/`；只 stage 该任务文件 + todo。

## 任务（纵切，按依赖顺序）
- [x] **R1** 配色令牌改造：金→陶土橙（--accent、price-ink 深陶土）+ 对比度测试 — S（地基）✅
- [ ] **R2** 首页融合 dashboard：墨绿头 + 余额/返利真数据 + 厨师彩色插画 banner + 入口卡 — L（依赖 R1）
- [ ] **R3** 菜品照片优先卡（product.image + 彩色展位图）+ 类目数量徽章 + 购物车带图 — L（依赖 R1）
- [ ] **R4** 下单复刻：商品明细带缩略图 + 支付方式（余额/微信）+ 支付成功态 — M（依赖 R3）
- [ ] **R5** 底部 tab 墨绿复刻 + PNG 图标（生成不足则请你提供）— M（依赖 R1）

## Checkpoint
- [ ] **C1** /frontend-design 自审 + 全量 jest 绿 + 真机核对（待人工：本环境无模拟器）

## 不在本期
- 优惠券 / 自提 / 预约时间（无后端）
- 真实菜品摄影（OSS 后填）
- 菜单内 自提/外卖 切换（M1 已删，单型在首页定）
- 地址 / 选店 / 我的订单屏
