# go-blog Admin

管理端（Vue 3 + Vite）。环境变量见 `.env`：`VITE_PORT=5174`、`VITE_API_PROXY_TARGET`。

路由资源来自 `GET /admin/user/resource/current`（静态菜单）。

Docker 生产构建注入 `VITE_BASE_URL=/admin/`、`VITE_API_URL=/api`。
