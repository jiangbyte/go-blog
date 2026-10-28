# go-blog frontend

pnpm workspace：

| 包 | 说明 | 端口 |
|----|------|------|
| `@jiangbyte/blog` | 博客前台（原 jiangbyte-web） | 5173 |
| `@jiangbyte/admin` | 管理端（自 mypage admin） | 5174 |
| `@jiangbyte/shared` | HTTP / RSA 等共享代码 | — |

```bash
pnpm install
pnpm dev:blog
pnpm dev:admin
```

根目录 `make blog` / `make admin` / `make dev` 更省事。
