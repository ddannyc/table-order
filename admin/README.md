# 商家管理后台 (admin)

Vue 3 + Element Plus + Vite 的商家管理后台 SPA。

## 本地开发

```bash
npm install
npm run dev      # http://localhost:5173 （读取 .env.development → 后端 localhost:8080）
npm test         # Vitest 单测
npm run build    # 产出 dist/
```

## 部署到 Cloudflare Pages

这是一个纯静态 SPA，推荐用 Cloudflare Pages（与 R2 / bestluckbox.com 同账号）。

**当前已上线**：https://table-order-admin.pages.dev （项目名 `table-order-admin`，production branch `main`）。

### 方式一：Git 集成（推荐）
在 Cloudflare Dashboard → Pages → Create project → 连接本仓库，按下表填**构建设置**：

| 设置项 | 值 |
|--------|-----|
| Production branch | `main`（或你的发布分支） |
| **Root directory** (Build → Advanced) | `admin` |
| Build command | `npm run build` |
| Build output directory | `dist` |
| Node 版本 | 20（环境变量 `NODE_VERSION=20`） |

**环境变量**（Pages → Settings → Environment variables，Production）：

| 变量 | 值 |
|------|-----|
| `VITE_API_BASE` | 后端公网地址，如 `https://<your-railway-app>.up.railway.app/api` |

> ⚠️ `VITE_*` 是**构建时**注入的，改了要重新部署。若不设此变量，则回退到 `admin/.env.production` 里的默认值。

### 方式二：Wrangler CLI（手动部署，实际使用的步骤）
```bash
# 1. 登录 Cloudflare（首次，会打开浏览器 OAuth）
npx wrangler login

# 2. 创建 Pages 项目（仅首次；新版 wrangler 不会在 deploy 时自动创建）
npx wrangler pages project create table-order-admin --production-branch main

# 3. 构建并部署（之后每次更新只需这两步）
npm run build
npx wrangler pages deploy dist --project-name table-order-admin
```

> 工作目录有未提交改动时 wrangler 会告警，加 `--commit-dirty=true` 可静默。

### SPA 回退
`public/_redirects`（`/* /index.html 200`）已配置，Pages 会把所有路径回退到 `index.html`，刷新子路由不会 404。

### 自定义域名
Pages → Custom domains 绑你的域名（需托管在 Cloudflare）。绑定后用该域名访问。

## ⚠️ 部署后须知

- **CORS**：后端当前 `Access-Control-Allow-Origin: *`，跨域可用。上线后建议把后端 CORS 收紧为只允许本后台的域名（见 ship review 的 ticket）。
- **后端地址**：确保 `VITE_API_BASE` 指向**生产后端**，且后端的 R2 等环境变量已在 Railway 配好。
