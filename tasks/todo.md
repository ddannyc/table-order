# Todo：实现 v6 设计稿（首页 / 菜品 / 下单 / Tab）

详见 `tasks/plan.md`。视觉真源：`docs/design/mockup-v6.png`。
一任务一提交，TDD（RED→GREEN→回归 jest）。本环境无模拟器，像素对齐靠 C1 真机核对。
基线：绝不 git add `frontend/config.js` / `.claude/` / `specs/`；只 stage 该任务文件 + todo。

## 任务（纵切，按依赖顺序）
- [x] **D1** 配色令牌对齐 v6（BRAND→#234B3A + --green-2）+ 对比度测试 — S（地基）✅
- [x] **D2** 首页重构到 v6：品牌头(移除余额/福利金) + 堂食/外卖分段(线描图标) + 扫码卡 + v1厨师 banner + 福利放送 — L（依赖 D1）✅
- [x] **D3** 菜品页到 v6：彩色饮品占位插画（匹配真实饮品分类）+ 购物车条已在 tab 之上 — L（依赖 D1）✅
- [x] **D4** 下单页到 v6：福利金抵扣开关(墨绿 switch) + 支付方式仅微信(绿徽+勾) + 缩略图(R4) + 移除冗余福利金卡 — M（依赖 D3）✅
- [x] **D5** Tab 线描图标对齐 v6（激活色 #2C4A3B→BRAND #234B3A，线描风格已符合）— S/M（依赖 D1）✅

## Checkpoint
- [ ] **C1** 全量 jest 绿 + 真机/开发者工具逐屏对照 mockup-v6.png（待人工：本环境无模拟器）

## Resolved Decisions（已定）
- 福利放送 = 静态装饰卡（不接数据、不可点、占位，日后可接精选菜品）
- 堂食/外卖分段：选外卖即 chooseDelivery 进配送流；扫码卡仅堂食 scanDineIn

## 不在本期
- 真实菜品摄影（OSS 后填）；优惠券/自提/预约时间（无后端）；地址/选店/我的订单屏
