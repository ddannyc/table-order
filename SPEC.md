# Spec: 微信原生小程序重构 (基于weui-wxss)

## Objective

将当前 uni-app 版本小程序改为**微信原生开发**，基于 `weui-wxss` 基础样式库，实现相同业务功能（扫码点餐、订单支付、邀请裂变、钱包）。

**用户**: 餐厅顾客 / 会员
**核心流程**: 扫码绑定桌号 → 浏览菜品 → 加购 → 确认支付 → 完成
**成功标准**: 所有现有页面功能保留，视觉升级为微信原生风格，构建产物可导入微信DevTools

---

## ASSUMPTIONS

1. Backend Go API (`localhost:8080`) — 不变
2. 业务逻辑 — 不变，仅前端技术栈切换
3. 页面保留: login, home, invite, profile, order-confirm
4. Tab顺序: 点餐 / 邀请 / 我的
5. 当前 `api/index.js` / `api/product.js` 的函数签名不变，仅替换 `uni.request` → `wx.request`
6. 存储从 `uni.getStorageSync` → `wx.getStorageSync`

---

## Tech Stack

| 层次 | 技术选型 | 说明 |
|------|----------|------|
| 框架 | 原生微信小程序 | `project.config.json` + `app.json` + 4文件结构 |
| 基础样式 | weui-wxss (npm) | 腾讯官方基础样式库，微信原生视觉 |
| 业务组件 | Vant Weapp (npm) | 按钮、Dialog、Toast、ActionSheet、Cell等 |
| 图标 | weui-icon / 内置icon | 避免第三方字体图标依赖 |
| 构建 | 微信DevTools内置编译 | 无需额外构建工具链 |
| 包管理 | npm → miniprogram_npm | 微信原生npm支持 |

**不采用 weapp-tailwindcss** — 引入额外学习成本，原子化CSS在小程序场景收益有限。

---

## Project Structure

```
frontend/
├── dist/mp-weixin/          # 构建输出，导入微信DevTools（当前已存在）
├── src/                     # 源码（开发时由微信DevTools直接引用src/）
│   ├── app.js               # 应用入口
│   ├── app.json             # 全局配置（pages/tabBar/window）
│   ├── app.wxss             # 全局样式（import weui-wxss + brand变量）
│   ├── pages/
│   │   ├── login/           # 登录页
│   │   │   ├── index.js
│   │   │   ├── index.wxml
│   │   │   └── index.wxss
│   │   ├── home/            # 点餐首页
│   │   ├── invite/          # 邀请页
│   │   ├── profile/         # 我的（含钱包+订单）
│   │   └── order-confirm/   # 订单确认+支付
│   ├── components/          # 业务组件
│   │   ├── product-card/    # 产品卡片
│   │   ├── cart-bar/        # 购物车悬浮栏
│   │   └── amount-panel/   # 金额展示面板
│   ├── api/                 # API客户端层
│   │   ├── index.js         # 店铺/用户/钱包/邀请 API
│   │   └── product.js       # 购物车/产品 API
│   └── utils/
│       └── storage.js       # Storage封装（wx.getStorageSync/setStorageSync）
├── miniprogram_npm/         # npm包构建产物（weui-wxss, vant-weapp）
├── package.json
└── project.config.json
```

---

## Commands

```bash
# 安装npm依赖（推荐）
npm install

# 或使用微信DevTools直接编译src/
#   微信DevTools → 导入项目 → 选择 frontend/
#   项目根目录设为 frontend/src/
```

---

## Code Style

### WXML 模板风格
```html
<view class="page">
  <view class="weui-form">
    <view class="weui-cells__group">
      <view class="weui-cell">
        <view class="weui-cell__bd">标题</view>
      </view>
    </view>
  </view>
</view>
```

### JS 模块风格（ES6）
```javascript
// api/index.js
const API_BASE = 'http://localhost:8080/api'

export const getShop = (shopId) => {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE}/shops/${shopId}`,
      method: 'GET',
      success: (res) => resolve(res.data),
      fail: reject
    })
  })
}

export const getTableBinding = () => {
  return {
    shopId: wx.getStorageSync('current_shop_id') || 0,
    tableNo: wx.getStorageSync('current_table_no') || ''
  }
}
```

### 全局样式变量（app.wxss）
```wxss
/* app.wxss */
@import 'miniprogram_npm/weui-wxss/dist/style/weui.wxss';

page {
  --weui-primary: #07c160;
  --weui-primary-light: rgba(7, 193, 96, 0.12);
}

page, view, text {
  font-family: -apple-system, BlinkMacSystemFont, sans-serif;
}
```

### 命名规范
- 组件目录: kebab-case (`product-card`)
- 样式类: 使用 `weui-` 前缀 + 业务类名
- JS变量: camelCase
- WXML: 2空格缩进

---

## Testing Strategy

- **手动测试** — 微信DevTools「编译」模式，实时预览
- **API验证** — 接口直接复用后端 Go API，确保数据流正确
- **重点验证**:
  1. 扫码 → 桌号绑定 → 存储
  2. 加购 → 购物车数量/金额联动
  3. 支付 → 余额/福利金抵扣 → 订单创建
  4. 邀请 → 生成邀请码 → 分享

---

## Boundaries

### Always
- 使用 `weui-wxss` 组件类名（`weui-cell`, `weui-btn` 等）确保原生视觉一致
- 品牌绿 `#07c160` 作为主色调通过 CSS 变量覆盖
- `wx.request` 封装在 `api/` 目录，对外保持相同函数签名
- Storage 统一封装在 `utils/storage.js`

### Ask First
- 添加新的 npm 依赖包
- 修改 `app.json` tabBar / pages 配置
- 修改 API 接口签名

### Never
- 不修改 `backend/` 代码（纯前端重构）
- 不删除 `tmp/` 测试二维码文件
- 不修改已有 API endpoint 路径

---

## Success Criteria

1. [ ] 微信DevTools可正常导入项目，无编译报错
2. [ ] TabBar 3页（点餐/邀请/我的）可正常切换
3. [ ] 扫码绑定桌号流程走通
4. [ ] 菜品展示、加购、购物车联动走通
5. [ ] 订单确认页余额/福利金展示正确
6. [ ] 支付流程走通
7. [ ] 邀请页邀请码生成/复制走通
8. [ ] 我的页面收支明细/订单列表展示正确
9. [ ] weui-wxss 基础样式正常渲染，无样式冲突
10. [ ] `tmp/` 目录下测试QR码保留

---

## Open Questions

1. 是否需要保留 H5 版本（当前 `npm run dev:h5`）？如需保留，同步维护一套还是放弃？
2. 登录页微信授权流程：当前是 auth code 换 openid，是否需要调整？
3. 是否需要接入 Vant Weapp 表单组件替代原生 input？