# Spec: 按需延迟登录 (Lazy Login)

## Objective

将微信小程序从"启动即强制登录"改为"按需延迟登录"。用户打开小程序后可正常浏览公开内容，在需要用户身份的功能区域显示"授权登录"提示，点击后引导登录。

**用户故事：**
- 作为新用户，打开小程序直接看到扫码绑桌、浏览菜单，不需要登录
- 作为新用户，点击 tab "邀请"可以进入邀请页，但邀请统计、邀请二维码等区域显示"登录后查看"
- 作为新用户，点击 tab "我的"可以进入个人页，但钱包、订单、福利金等区域显示"登录后查看"
- 作为新用户，加完购物车点击"去结算"，跳转到订单确认页，但下单支付按钮区域显示"登录后下单"
- 作为已登录用户，体验与之前完全一致

## Current State

### 当前登录流程

```
App.onLaunch
  → 检查 token
  → 无 token → wx.reLaunch('/pages/login/index')  ← 强制拦截所有页面
```

### 后端路由权限

| 路由 | 需要登录 |
|------|----------|
| `GET /api/shops/:id` | No |
| `GET /api/shops/:id/products` | No |
| `GET /api/auth/login` | No |
| 其他所有 `/api/*` | **Yes** |

### 前端页面权限（目标）

| 页面 | 目标 | 未登录时展示 |
|------|------|-------------|
| 首页 (点餐) | **无需登录** | 扫码绑桌、菜单、加购物车全部可用 |
| 订单确认 | **可进入** | 购物车列表可见，下单按钮区域显示"登录后下单" |
| 邀请页 | **可进入** | 邀请统计/二维码区域显示"登录后查看" |
| 我的 | **可进入** | 钱包/订单/福利金区域显示"登录后查看" |
| 分享码 | **可进入** | 二维码/邀请链接区域显示"登录后查看" |
| 登录页 | 登录中间页 | 登录后返回来源页 |

## Design

### 核心改动

1. **`app.js` onLaunch** — 删除 token 检查 + reLaunch 登录。只处理 scene 参数。
2. **API 层 401** — 清除 token，不再 reLaunch 登录页。由页面逻辑决定何时跳转。
3. **`utils/auth.js`** — 新增守卫函数 `requireLogin()`，无 token 时跳转登录页，登录后返回来源页。
4. **各页面** — 需要用户身份的 API 调用 catch 到 401 时，将对应区域切换为"登录提示"视图。
5. **登录页** — 登录成功后优先返回来源页（`return_path`），而非固定跳首页。

### requireLogin() 守卫

```js
// utils/auth.js
function requireLogin() {
  const token = wx.getStorageSync('token')
  if (token) return true
  // 记录当前页面路径，登录后返回
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1]
  if (currentPage) {
    const route = '/' + currentPage.route
    const options = currentPage.options || {}
    const query = Object.keys(options).map(k => k + '=' + options[k]).join('&')
    wx.setStorageSync('return_path', route + (query ? '?' + query : ''))
  }
  wx.navigateTo({ url: '/pages/login/index' })
  return false
}
```

### 页面改造模式

每个需要用户身份的页面采用"降级展示"模式，而非"强制跳转"：

```
onLoad / onShow:
  读取 token
  有 token → 调用 API 加载用户数据
  无 token → 设置 data.needLogin = true，展示登录提示区域

用户点击"登录"按钮 → 调用 requireLogin()
  登录成功 → 回到本页 → onLoad/onShow 重新执行 → 此时有 token → 正常加载
```

### 登录页改动

登录成功后：
1. 读取 `return_path`，有则 `wx.redirectTo` 回去
2. 无则 `wx.reLaunch` 到首页（兼容旧行为）

### 401 处理改动

`api/index.js` 的 401 处理：
- 清除 token
- reject Promise（由页面 catch 处理，展示登录提示）
- **不再** `wx.reLaunch` 到登录页

## Success Criteria

- [ ] 无 token 用户打开小程序，直接看到扫码绑桌引导页
- [ ] 无 token 用户扫码后可浏览菜单、加购物车
- [ ] 无 token 用户点击 tab "邀请" → 进入邀请页，看到"登录后查看邀请统计"提示
- [ ] 无 token 用户点击 tab "我的" → 进入个人页，钱包/订单/福利金区域显示登录提示
- [ ] 无 token 用户点击"去结算" → 进入订单确认页，购物车可见，下单按钮区域显示"登录后下单"
- [ ] 无 token 用户点击任一登录提示 → 跳转登录页 → 登录后回到原页面 → 正常展示内容
- [ ] 已登录用户体验不变
- [ ] token 过期时 API 返回 401 → 页面对应区域降级为登录提示
- [ ] 扫码邀请码场景：未登录用户扫码 → 存 pending_invite_code → 登录后自动绑定

## Files Changed

```
frontend/app.js                          → 删除 token 检查
frontend/utils/auth.js                   → 新增 requireLogin() + return_path 逻辑
frontend/api/index.js                    → 401 改为清除 token + reject，不 reLaunch
frontend/pages/login/index.js            → 登录成功后返回 return_path
frontend/pages/login/index.js            → 登录页 onLoad 有 token 时跳 return_path
frontend/pages/home/index.js             → 无需改动（API 调用失败 catch 即可）
frontend/pages/invite/index.js           → 无 token 时降级展示，点击登录提示调 requireLogin
frontend/pages/profile/index.js          → 无 token 时降级展示，点击登录提示调 requireLogin
frontend/pages/order-confirm/index.js    → 无 token 时降级展示，点击"下单"调 requireLogin
frontend/pages/share-code/index.js       → 无 token 时降级展示，点击登录提示调 requireLogin
```

## Boundaries

- **Always do:** 保持后端 API 权限不变
- **Ask first:** 降级展示的 UI 样式（登录提示按钮的文案、样式）
- **Never do:** 把需要登录的 API 改成公开接口；在 app.js 里加任何形式的登录拦截
