# jb-md

本机 Markdown 上传 CLI：在开发机读本地 `.md` / 图片，登录后写入**正式站**博客草稿（默认 API：`https://example.com/api`）。

CLI 本身不进 Docker；需正式站已部署含 CLI 所需接口的 API（登录回传 token、`posts/by-title` 等）。

## 构建

在仓库根目录：

```bash
make jb-md          # → bin/jb-md
# 或
go build -o bin/jb-md ./cmd/jb-md
go install ./cmd/jb-md
```

产出为单文件可执行（登录页已 `go:embed`），可直接拷贝使用。

## 用法

```bash
./bin/jb-md login                                    # 浏览器滑块登录正式站，凭证落盘
./bin/jb-md upload note.md
./bin/jb-md upload --dir ~/Notes --glob '**/*.md'
./bin/jb-md upload --dry-run ./a.md
./bin/jb-md upload --on-exists overwrite ./a.md
./bin/jb-md logout
```

连本机 API（可选）：

```bash
./bin/jb-md login --base http://127.0.0.1:8000
# 或 export JB_API_BASE=http://127.0.0.1:8000
```

- 凭证：`~/.config/go-blog/credentials.json`（`0600`）
- 优先级：`--token` / `JB_TOKEN` > 文件；`--base` / `JB_API_BASE` > 文件 > 默认正式站
- 分类默认「其它」（不存在则自动创建）；标签为空
- 同标题默认交互：`o` 覆盖 / `s` 跳过 / `c` 新建带后缀标题
