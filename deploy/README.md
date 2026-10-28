# 部署（Docker）

本目录集中管理打包与上线：镜像 Dockerfile、Compose 编排、环境变量样例。

网关使用 **Caddy**：未配置域名时仅 HTTP；配置 `SITE_DOMAIN` 后自动 Let's Encrypt HTTPS。

```
deploy/
├── README.md
├── docker-compose.yml          # 本地/自建构建
├── docker-compose.prod.yml     # 生产：只拉镜像
├── .env.example                # 复制为 .env 后填写（勿提交 .env）
├── scripts/docker-build-push.sh
└── docker/
    ├── api/Dockerfile
    ├── blog/Dockerfile
    ├── admin/Dockerfile
    ├── resume/Dockerfile
    └── gateway/                # Caddy：HTTP / 自动 HTTPS
```

构建上下文为**仓库根目录**（`context: ..`），根目录 `.dockerignore` 生效。

## 依赖（Compose 不内置）

| 服务 | 说明 |
|------|------|
| MySQL | 默认 `host.docker.internal:3306`，库 `go_blog` |
| Redis | 默认 `host.docker.internal:6379` |
| 七牛 | `.env` 填 `STORAGE_QINIU_*`，`DEFAULT_FILE_ENGINE=QINIU` |

建库：在仓库根执行 `make sql`（经本机 Docker mysql）。

## 本地启动（HTTP）

```bash
cd deploy
cp .env.example .env          # 填 MYSQL_*、REDIS_*、七牛 AK/SK 等
# 不要设 SITE_DOMAIN → 仅 HTTP（默认映射 8088）
docker compose up -d --build
```

或在仓库根：`make docker-up` / `make docker-down` / `make docker-logs`。

| 入口 | 地址 |
|------|------|
| Blog | http://127.0.0.1:8088/ |
| Admin | http://127.0.0.1:8088/admin/ |
| API（经网关） | http://127.0.0.1:8088/api/health |
| API（本机直连） | http://127.0.0.1:8000/health |

## 线上部署（拉镜像）

服务器**不构建**，只拉仓库里的 `:latest`。本机先推镜像：

```bash
# 按需覆盖 registry / namespace
export REGISTRY=your-registry.example.com
export NAMESPACE=your-namespace
make docker-push IMAGE_VERSION=1.0.0
```

### 1. 前置

| 项 | 要求 |
|----|------|
| DNS | `example.com` / `www` / `resume.example.com` A → 服务器公网 IP（或 CNAME → apex；**不要**指对象存储） |
| 端口 | 放行 **80 / 443**（勿对公网开 8000 / MySQL / Redis） |
| MySQL / Redis | 已就绪，库 `go_blog` 已执行 `docs/sql` |
| 七牛 | AK/SK、Bucket、CDN 域名（如 `images.example.com`）已配好 |

### 2. 服务器上准备文件

只需 `deploy/` 里这两样（不必整仓源码）：

```bash
mkdir -p ~/go-blog/deploy && cd ~/go-blog/deploy
# 上传 docker-compose.prod.yml + .env（由 .env.example 改）
```

`.env` 生产关键项示例：

```bash
MYSQL_HOST=...          # 服务器能访问的 MySQL（同机可用 host.docker.internal）
REDIS_HOST=...
DEBUG=false
TOKEN_COOKIE_SECURE=auto
CORS_ALLOWED_ORIGINS=https://example.com,https://www.example.com,https://resume.example.com
PUBLIC_SITE_URL=https://example.com
SITE_DOMAIN=example.com
ACME_EMAIL=admin@example.com
RESUME_DOMAIN=resume.example.com
GATEWAY_HTTP_PORT=80
GATEWAY_HTTPS_PORT=443
# 七牛 STORAGE_QINIU_* 填齐；IMAGE_DOMAIN / IMAGE_UPSTREAM 见下文
```

将 `docker-compose.prod.yml` 中的镜像前缀改成你的 `REGISTRY/NAMESPACE`（默认占位 `your-registry.example.com/your-namespace/...`）。

### 3. 登录仓库并启动

```bash
docker login your-registry.example.com
cd ~/go-blog/deploy
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d --force-recreate
docker compose -f docker-compose.prod.yml ps
curl -sS https://example.com/api/health
```

| 端 | 地址 |
|----|------|
| Blog | https://example.com/ |
| Admin | https://example.com/admin/ |
| Resume | https://resume.example.com/ （公开 PDF；需 Admin 先发布简历） |
| API | https://example.com/api/health |
| 图床 | https://images.example.com （可选：Caddy 终结 TLS，回源七牛 HTTP） |

### 简历子域名

1. DNS：`resume.example.com` A → 服务器 IP（或 CNAME → apex），**不要**指七牛
2. `.env` 设 `RESUME_DOMAIN=resume.example.com`，CORS 含该 Origin
3. Admin 发布一份公开简历并生成 PDF
4. `pull` + `up -d --force-recreate` gateway / resume
5. 若 HTTPS 握手异常：确认 DNS 生效后清理失败证书缓存再启 gateway：

```bash
cd ~/go-blog/deploy
docker compose -f docker-compose.prod.yml stop gateway
docker compose -f docker-compose.prod.yml run --rm --entrypoint sh gateway -c \
  'rm -rf /data/caddy/certificates/*/*resume.example.com* /data/caddy/acme/*/*resume* 2>/dev/null; ls /data/caddy/certificates 2>/dev/null || true'
docker compose -f docker-compose.prod.yml up -d gateway
docker compose -f docker-compose.prod.yml logs -f --tail=80 gateway
curl -sSI https://resume.example.com/
```

### 图床 HTTPS（可选）

若希望浏览器走 HTTPS、对象存储侧不开 HTTPS 计费：由本机 Caddy 终结 TLS，再 **HTTP 回源** CDN。

1. DNS：把 `images.example.com` 指到本机（A 记录）
2. 记下 CDN 的回源主机名，填入 `IMAGE_UPSTREAM`（如 `cdn.example.com`）
3. `.env`：

```bash
IMAGE_DOMAIN=images.example.com
IMAGE_UPSTREAM=cdn.example.com
STORAGE_QINIU_USE_SSL=true
STORAGE_QINIU_BASE_URL=https://images.example.com
```

4. 重建并推送 gateway 后，服务器 `pull` + `up -d --force-recreate gateway api`

流量：`浏览器 HTTPS → Caddy → HTTP → 对象存储 CDN`。

### 4. 升级发版

```bash
# 本机：打版本号 + 覆盖 latest
make docker-push IMAGE_VERSION=1.0.0
# 服务器：只拉 latest
cd ~/go-blog/deploy
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d --force-recreate
```

### 常用运维

```bash
docker compose -f docker-compose.prod.yml logs -f --tail=200 gateway
docker compose -f docker-compose.prod.yml logs -f --tail=200 api
docker compose -f docker-compose.prod.yml down
docker volume ls | grep go-blog   # api_data / gateway_caddy_data 勿误删
```

首次登录后台后立刻改掉默认 `admin` / `admin123`。

## 推送镜像

```bash
REGISTRY=your-registry.example.com NAMESPACE=your-namespace \
  ./deploy/scripts/docker-build-push.sh 1.0.0
# 或：make docker-push IMAGE_VERSION=1.0.0
```

| 镜像 | 仓库名 |
|------|--------|
| api | `…/go-blog-api` |
| blog | `…/go-blog-blog` |
| admin | `…/go-blog-admin` |
| resume | `…/go-blog-resume` |
| gateway | `…/go-blog-gateway` |

推送时同时打 `:VERSION` 与 `:latest`；线上 `docker-compose.prod.yml` 固定用 `:latest`。
