# Todo：/ship 评审阻断项修复（外卖/闪送支付链路）

详见 `tasks/plan.md`。一任务一提交，TDD（RED→GREEN→回归）。
基线：绝不 git add `frontend/config.js` / `.claude/` / `specs/shansong-delivery.md`；密钥仅 env。

## 阻断项
- [ ] **BLK1** 报价密钥 fail-closed（去常量回退）+ 拒绝负运费 — S
- [ ] **BLK2** 暴露派单状态：0=待派单 / -1=派单失败 落库与标签 — S
- [ ] **BLK3** 支付转移原子化（守卫 UPDATE，竞态只发一次奖励/派单）— M

## 推荐项
- [ ] **BLK4** 回调禁止状态回退（终态 50/60 不被覆盖）— S

## Checkpoint
- [ ] 三套测试全绿（go / jest 94 / vitest 21）
- [ ] go build + vet 通过；无真实密钥；未 add 基线三文件

## 不在本次范围（后续）
- 派单失败的自动重试 / 运营告警 / 重派路由
- 报价令牌绑定 user_id + 单次使用（当前靠 order_no 唯一索引兜底）
- 回调 timestamp 新鲜度窗口
- 上游 README/changelog；真机全链路 + 回调 wrapper 联调
