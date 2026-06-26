# Todo：质感升级 首页+菜单（松墨 Pine-Ink）

详见 `tasks/plan.md`。来源 spec：`docs/ideas/texture-uplift-pine-ink.md`。
一任务一提交，TDD（RED→GREEN→回归 jest）。
基线：绝不 git add `frontend/config.js` / `.claude/` / `specs/`；只 stage 该任务文件 + todo。

## 任务（纵切，按依赖顺序）
- [x] **T1** 令牌地基：app.wxss 覆盖 weui 配色 + accent/price + WCAG 对比度测试 — S（地基）✅
- [ ] **T2** 首页品牌 band + 入口卡重皮 — M（依赖 T1）
- [ ] **T3** 首页 hero 线描插画（蒸笼一桌菜，data-URI SVG）— M（依赖 T2）
- [ ] **T4** 菜单重皮（shopbar/rail/cards/price/cartbar 令牌上色）— M（依赖 T1）
- [ ] **T5** 菜单分类金线 glyph（升级 M3 占位）— M（依赖 T1、T4）
- [ ] **T6** 未绑桌空状态线描插画 — S（依赖 T4）

## Checkpoint
- [ ] **C1** /frontend-design 自审 + 全量 jest 绿 + 真机核对（待人工：本环境无模拟器）

## 不在本期
- 真实食物摄影 / 后台图片上传
- 其余 4 屏（下单/我的订单/地址/选店）重皮
- 自定义衬线/显示字体
- 类目 scroll-spy / 锚点联动
- 余额 / 会员积分模块上首页
