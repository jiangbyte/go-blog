# go-blog

个人博客全栈模板：**Go（Gin + GORM）** + **Vue 前台 / 管理端 / 简历站**。单管理员、弱化 RBAC，开箱可跑本地开发与 Docker 部署。

## 特性

- 文章、分类 / 标签、项目、归档、搜索、关于页
- 管理端 Markdown 编辑、站点配置、文件上传（默认七牛 Kodo）
- 简历：管理端编辑 + 公开 HTML / PDF（PDF 依赖 Chromium）
- 滑块验证码登录；响应信封 `{ code, message, data }`（成功 `0000`）
- URL 约定：`/auth`、`/admin/...`、`/portal/...`；仅 GET / POST；JSON **snake_case**
- 可选 sitemap / IndexNow 推送

## 目录结构

```
cmd/server/          API 入口
cmd/jb-md/           Markdown 上传 CLI
internal/            config / auth / blog / resume / file / storage …
docs/sql/            MySQL schema + seed
deploy/              Docker Compose、Caddy 网关、镜像构建
web/apps/blog/       前台（Vite :5173）
web/apps/admin/      管理端（Vite :5174）
web/apps/resume/     简历站（Vite :5175）
web/packages/        shared / markdown
```

## 环境依赖

| 依赖 | 说明 |
|------|------|
| Go 1.22+ | 后端 |
| Node 20+ / pnpm | 前端 monorepo |
| Docker | 本机 MySQL、Redis（本仓库**不**自起这两个容器） |
| 七牛（可选） | 对象存储；也可改回 MinIO（见 `.env.example` 注释） |
| Chromium（可选） | 简历 PDF 导出；Docker API 镜像已内置 |

默认约定本机已有：

| 容器 | 端口 | 示例凭证 |
|------|------|----------|
| `mysql` | 3306 | root / `infra123!`，库 `go_blog` |
| `redis` | 6379 | 密码 `infra123!` |

## 快速开始（本机开发）

```bash
git clone https://github.com/jiangbyte/go-blog.git
cd go-blog

make env          # 复制 .env.example → .env
# 按需编辑 .env：MySQL / Redis / 七牛 STORAGE_QINIU_* 等

make install      # go mod download + pnpm install
make sql          # 建库建表 + seed（经 docker exec mysql）
make dev          # 同时起 api + blog + admin + resume
```

默认管理员：`admin` / `admin123`（上线后请立刻修改）。

| 端 | 地址 |
|----|------|
| Blog | http://127.0.0.1:5173 |
| Admin | http://127.0.0.1:5174 |
| Resume | http://127.0.0.1:5175 |
| API | http://127.0.0.1:8000/health |

也可单独：`make api` / `make blog` / `make admin` / `make resume`。

## 配置说明

根目录 [`.env.example`](.env.example) 为开发模板，复制为 `.env` 后填写。常用项：

| 变量 | 作用 |
|------|------|
| `MYSQL_*` / `REDIS_*` | 数据库与缓存 |
| `PUBLIC_SITE_URL` | 对外站点根（sitemap 等），示例 `https://example.com` |
| `CORS_ALLOWED_ORIGINS` | 允许的前端 Origin（逗号分隔） |
| `DEFAULT_FILE_ENGINE` | 默认 `QINIU`；需填 `STORAGE_QINIU_*` |
| `STORAGE_QINIU_BASE_URL` | CDN / 图床访问前缀 |
| `INDEXNOW_KEY` | 空则关闭 IndexNow；非空时需在 `web/apps/blog/public/{key}.txt` 放同名验证文件（内容为密钥，勿提交真实密钥） |
| `AUTH_RSA_KEY_PATH` | RSA 私钥路径；首次运行可自动生成到 `data/`（已 gitignore） |
| `TOKEN_COOKIE_SECURE` | `true` / `false` / `auto` |

Docker 部署另见 [`deploy/.env.example`](deploy/.env.example)。

## Docker 部署

编排与镜像在 **`deploy/`**，网关为 **Caddy**：未设 `SITE_DOMAIN` 时仅 HTTP；生产设 `SITE_DOMAIN=example.com` 可自动 HTTPS。

```bash
cd deploy
cp .env.example .env    # 填 MySQL / Redis / 七牛等
docker compose up -d --build
# 或仓库根：make docker-up
```

本地网关入口示例：`http://127.0.0.1:8088/` · `/admin/` · `/api/health`。

更完整的生产 DNS、图床回源、简历子域、镜像推送说明见 **[deploy/README.md](deploy/README.md)**。

## 常用 Make 命令

| 命令 | 说明 |
|------|------|
| `make env` | 生成 `.env` |
| `make install` | 安装前后端依赖 |
| `make sql` | 按序执行 `docs/sql/*.sql` |
| `make sql-list` / `make sql-one F=…` | 列出 / 执行单个 SQL |
| `make dev` | 本地四端开发 |
| `make build` | 构建 API 与前端 |
| `make jb-md` | 构建 Markdown 上传 CLI → `bin/jb-md` |
| `make docker-up` / `docker-down` / `docker-logs` | Compose 启停与日志 |
| `make docker-push IMAGE_VERSION=1.0.1` | 构建并推送镜像（需配置 registry） |

`jb-md` 用法见 [`cmd/jb-md/README.md`](cmd/jb-md/README.md)。

## API（节选）

- `GET /auth/public-key` · `POST /auth/login` · `GET /auth/me` · `POST /auth/logout`
- `GET /admin/user/resource/current`（静态菜单，无资源表）
- `GET /portal/blog/site` · `posts/*` · `projects/*` · `archive/*` · `taxonomy/*` · `search` · `about`
- 简历：`/admin/resume/resume/*` · `GET /portal/resume/resume` · `/html`
- 文件：`/admin/sys/file/*`（元数据表 `sys_file`）

Blog 前端经 Vite 代理 `/api` → Gin。前台路由含 `/resume`。

## 说明

- 单管理员：不做多用户 / 角色树；登录即为全权限。
- Seed 仅为演示内容，可按需在管理端改掉。
- 本模板已脱敏：请将 `example.com`、空密钥等换成你自己的配置，**不要**把 `.env`、`data/`、数据库 dump 提交到公开仓库。
