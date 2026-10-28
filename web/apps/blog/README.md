# go-blog（前台）

Demo Blog主题的 Vue 3 站点（Vite + TypeScript + Tailwind CSS）。数据经 `/api` 对接 Gin 后端。

## 开发

```bash
npm install
npm run dev
```

需本地后端（默认 `VITE_API_PROXY_TARGET=http://127.0.0.1:8000`）。

```bash
npm run build
npm run preview
```

## 路由

| Path | 说明 |
|------|------|
| `/` | 首页（Banner + 文章流 + 侧栏） |
| `/p/:uid` | 文章详情 |
| `/projects` | 项目列表 |
| `/projects/:uid` | 项目详情 |
| `/archives` | 归档 |
| `/categories` `/categories/:uid` | 分类 |
| `/tags` `/tags/:uid` | 标签 |
| `/project-tags` `/project-tags/:uid` | 项目标签 |
| `/about` | 关于 |
| 其它 | 404 |

API 封装见 `src/api/`。

## 说明

- 主题色与 CSS 变量见 `src/styles/theme.css`（仅亮色）
- 图标：`@icon-park/vue-next`，统一经 `AppIcon` 使用
