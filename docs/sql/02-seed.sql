USE go_blog;

-- admin / admin123（头像用站点占位图，避免前台作者卡空白）
INSERT INTO admin_user (id, username, password_hash, nickname, avatar, signature, user_type, enabled) VALUES
  (1, 'admin', '$2a$10$liah34hDbqYmH5BSeHEihOviO9FsB4nd56reyGKy9OJnNz.6YS0I.', 'Admin', '/img/logo.png', 'Demo Blog 管理员', 'ADMIN', 1)
ON DUPLICATE KEY UPDATE
  password_hash = VALUES(password_hash),
  nickname = VALUES(nickname),
  avatar = VALUES(avatar),
  signature = VALUES(signature);

INSERT INTO blog_site (id, config_json) VALUES (1, JSON_OBJECT(
  'title', 'Demo Blog',
  'description', '个人笔记与实践记录 · Demo Blog',
  'site_name', 'Demo Blog',
  'icon', '/img/logo.png',
  'banner', JSON_OBJECT('title', 'Demo Blog', 'desc', '笔记与实践记录'),
  'nav', JSON_ARRAY(
    JSON_OBJECT('name', '文库', 'children', JSON_ARRAY(
      JSON_OBJECT('name', '全部文章', 'href', '/archives', 'icon', 'folder'),
      JSON_OBJECT('name', '全部分类', 'href', '/categories', 'icon', 'clone'),
      JSON_OBJECT('name', '全部标签', 'href', '/tags', 'icon', 'tags')
    )),
    JSON_OBJECT('name', '项目', 'children', JSON_ARRAY(
      JSON_OBJECT('name', '全部项目', 'href', '/projects', 'icon', 'folder'),
      JSON_OBJECT('name', '项目标签', 'href', '/project-tags', 'icon', 'tags')
    )),
    JSON_OBJECT('name', 'GitHub', 'href', 'https://github.com/example', 'icon', 'github', 'external', true),
    JSON_OBJECT('name', '关于', 'href', '/about', 'icon', 'user')
  ),
  'nav_group', JSON_ARRAY(
    JSON_OBJECT('name', 'GitHub', 'url', 'https://github.com/example', 'icon', 'github'),
    JSON_OBJECT('name', 'Resume', 'url', 'https://resume.example.com', 'icon', 'file')
  ),
  -- 作者昵称/头像/签名由 admin_user（个人中心）注入；此处仅侧栏展示壳
  'author', JSON_OBJECT(
    'background_url', '/img/default.png',
    'witty_words', JSON_ARRAY('写清楚，再优化', '先跑通，再抽象'),
    'links', JSON_ARRAY(
      JSON_OBJECT('name', 'GitHub', 'url', 'https://github.com/example', 'icon', 'github'),
      JSON_OBJECT('name', 'Resume', 'url', 'https://resume.example.com', 'icon', 'file')
    )
  ),
  'default_covers', JSON_ARRAY(
    '/img/default.png'
  ),
  'started_at', '2024-10-01T00:00:00+08:00',
  'footer', JSON_OBJECT(
    'beian', JSON_ARRAY(JSON_OBJECT('name', '粤ICP备xxxxxxxx号', 'url', 'https://beian.miit.gov.cn/')),
    'copyright', '© 2026 Demo Blog'
  ),
  'page_size', 12
)) ON DUPLICATE KEY UPDATE config_json = VALUES(config_json);

INSERT INTO blog_about (id, doc_json) VALUES (1, JSON_OBJECT(
  'title', '关于',
  'content', CONCAT(
    '你好 👋', CHAR(10), CHAR(10),
    '这里是 **Demo Blog** 的示例关于页。', CHAR(10), CHAR(10),
    '可在管理端修改文案、导航与站点配置。', CHAR(10), CHAR(10),
    '---', CHAR(10), CHAR(10),
    '### 座右铭', CHAR(10), CHAR(10),
    '先跑通，再抽象', CHAR(10), CHAR(10),
    '---', CHAR(10), CHAR(10),
    '## Demo Blog', CHAR(10), CHAR(10),
    '本仓库自带演示内容，方便本地启动后立刻看到前台效果。'
  )
)) ON DUPLICATE KEY UPDATE doc_json = VALUES(doc_json);

-- taxonomy：categories / tags / project_tags（JSON 里只存 uid）
DELETE FROM blog_taxonomy;
INSERT INTO blog_taxonomy (uid, kind, name) VALUES
-- categories
('01a128b0426f7de88e52e2bb7b2ef905', 'category', '运维'),
('01a128b0426f7e039bf2ca5b52f27298', 'category', '其它'),
('01a128b0426f7e06ab33923cdcf5b5f5', 'category', '工具'),
('01a128b0426f7e0798c0bd1704698564', 'category', '实践'),
-- tags
('01a128b0426f7e10a3968788bbed7cf9', 'tag', 'Markdown'),
('01a128b0426f7e12bd141234ab8c88cf', 'tag', '测试'),
('01a128b0426f7e14a8353b52d7ccf082', 'tag', 'Shortcode'),
('01a128b0426f7e22825b66625e34e440', 'tag', 'Docker'),
('01a128b0426f7e24be2156ac079eaf80', 'tag', 'Redis'),
('01b001demo0000000000000000000001', 'tag', 'Go'),
('01b001demo0000000000000000000002', 'tag', 'Gin'),
('01b001demo0000000000000000000003', 'tag', 'Vue'),
('01b001demo0000000000000000000004', 'tag', 'API'),
('01b001demo0000000000000000000005', 'tag', '部署'),
-- project_tags
('01a128b0426f7e2ca83409ec8c567294', 'project_tag', 'Docker'),
('01a128b0426f7e2eadd749e3f8348586', 'project_tag', 'CLI'),
('01a128b0426f7e3091dce3365ceda7e4', 'project_tag', '前端'),
('01a128b0426f7e32b845844bb54c2a33', 'project_tag', '工具'),
('01b003demo0000000000000000000001', 'project_tag', 'Go'),
('01b003demo0000000000000000000002', 'project_tag', 'Vue'),
('01b003demo0000000000000000000003', 'project_tag', 'Demo');

DELETE FROM blog_post;
INSERT INTO blog_post (uid, title, description, published_at, updated_at, cover, categories_json, tags_json, reading_time_minutes, word_count, locate, body) VALUES
('01a128a9a64073408ce8976b098575fd', '01-Markdown特性完整测试',
 '覆盖 CommonMark / GFM / 代码高亮 / Mermaid / KaTeX / 增强 HTML 组件的厨房水槽测试页，用于视觉与构建回归。',
 '2026-10-02', NULL, '/img/default.png',
 '["01a128b0426f7e039bf2ca5b52f27298"]',
 '["01a128b0426f7e10a3968788bbed7cf9","01a128b0426f7e12bd141234ab8c88cf","01a128b0426f7e14a8353b52d7ccf082"]',
 18, 4200, 'Demo',
 '完整 Markdown 厨房水槽正文由 make sql 从 docs/seed/01-markdown-kitchen-sink.md 写入。\n'),
('01b002demo0000000000000000000001', '从零跑通 Gin + Vue 博客的本地开发',
 '用 Make 把 API、前台与管理端一次拉起：环境变量、依赖安装、SQL seed 与热更新联调要点。',
 '2026-09-18', '2026-09-18', '/img/default.png',
 '["01a128b0426f7e0798c0bd1704698564"]',
 '["01b001demo0000000000000000000001","01b001demo0000000000000000000002","01b001demo0000000000000000000003"]',
 8, 1600, 'Demo',
 '## 目标\n\n本地同时跑通 API（Gin）与 Blog / Admin（Vite），验证登录与文章列表。\n\n## 步骤\n\n1. make env 生成 .env，确认 MySQL / Redis 可连\n2. make install 拉 Go 与 pnpm 依赖\n3. make sql 建库并写入演示数据\n4. make dev 启动四端（或只起 api + blog）\n\n## 验收\n\n- GET /health 返回 code=0000\n- 前台首页能看到 seed 文章\n- 管理端可用默认账号登录（上线请立刻改密）\n\n## 备注\n\n对象存储可先空着；不上传文件时不影响读文章与站点配置。\n'),
('01b002demo0000000000000000000002', 'API 信封约定：code / message / data',
 '统一成功码 0000、snake_case 字段与仅 GET/POST 的接口风格，方便前后端与 CLI 共用同一套约定。',
 '2026-09-25', NULL, '/img/default.png',
 '["01a128b0426f7e06ab33923cdcf5b5f5"]',
 '["01b001demo0000000000000000000001","01b001demo0000000000000000000004"]',
 6, 1200, 'Demo',
 '## 为什么要信封\n\n业务错误与 HTTP 状态解耦：网络层 200 仍可通过 code 表达失败，前端拦截器只认一套规则。\n\n## 形状\n\n```json\n{\n  \"code\": \"0000\",\n  \"message\": \"ok\",\n  \"data\": {}\n}\n```\n\n## 约定摘要\n\n- 成功：code 为 0000\n- JSON 字段一律 snake_case\n- 路径分区：/auth、/admin/...、/portal/...\n- 只使用 GET / POST，复杂查询放 query 或 body\n\n## 联调提示\n\n管理端与 jb-md CLI 都按同一信封解析；改码表时记得同步前端拦截器。\n'),
('01b002demo0000000000000000000003', 'Docker Compose 网关：Caddy 反代 Blog / Admin / API',
 '本地用 8088 入口统一反代，生产再挂 SITE_DOMAIN 自动 HTTPS；admin 固定挂在 /admin/。',
 '2026-10-05', '2026-10-06', '/img/default.png',
 '["01a128b0426f7de88e52e2bb7b2ef905"]',
 '["01a128b0426f7e22825b66625e34e440","01b001demo0000000000000000000005"]',
 7, 1400, 'Demo',
 '## 结构\n\ndeploy/ 下用 Compose 编排 api、blog、admin、resume 与 gateway（Caddy）。\n\n## 本地\n\n```bash\ncd deploy\ncp .env.example .env\ndocker compose up -d --build\n```\n\n访问 http://127.0.0.1:8088/ 、/admin/ 、/api/health。\n\n## 生产要点\n\n- 设置 SITE_DOMAIN 后 Caddy 申请证书\n- MySQL / Redis 外置，Compose 不内置数据库\n- 图床可走独立域名并由网关回源\n\n更细的 DNS 与发版步骤见仓库 deploy/README.md。\n'),
('01b002demo0000000000000000000004', '用 Redis 扛住访问计数的热路径',
 '页面浏览先写 Redis，再异步落库；启动时可回填，避免每次打开文章都打 MySQL。',
 '2026-10-12', NULL, '/img/default.png',
 '["01a128b0426f7e0798c0bd1704698564"]',
 '["01a128b0426f7e24be2156ac079eaf80","01b001demo0000000000000000000001"]',
 5, 1100, 'Demo',
 '## 思路\n\n访问统计是典型的高写低读延迟敏感路径：先 INCR Redis，定时或批量同步到表。\n\n## 收益\n\n- 削峰：瞬时流量不直接压 MySQL\n- 前台读热数据更快\n- 进程重启后可从库回填 Redis\n\n## 注意\n\n同步任务要幂等；键命名带上内容 uid，避免串号。本地开发记得在 .env 配好 REDIS_PASSWORD。\n');

DELETE FROM blog_project;
INSERT INTO blog_project (uid, title, description, published_at, weight, cover, git, site, tags_json, body) VALUES
('01b003proj0000000000000000000001', 'Demo Portal',
 'Vue 前台演示：站点配置、文章流与项目列表联调用。',
 '2026-09-01', 10, '/img/default.png',
 'https://github.com/example/demo-portal', 'https://example.com',
 '["01b003demo0000000000000000000002","01a128b0426f7e3091dce3365ceda7e4","01b003demo0000000000000000000003"]',
 '## Demo Portal\n\n用于展示博客前台布局与公开 API 联调，内容均为占位数据。\n\n### 能力\n\n- 首页文章流与侧栏\n- 项目列表与详情\n- 站点导航 / 关于页\n'),
('01b003proj0000000000000000000002', 'Config Lint CLI',
 '检查 .env 与 Compose 关键项是否齐全的示例命令行工具。',
 '2026-09-20', 20, '/img/default.png',
 'https://github.com/example/config-lint-cli', NULL,
 '["01b003demo0000000000000000000001","01a128b0426f7e2eadd749e3f8348586","01b003demo0000000000000000000003"]',
 '## Config Lint CLI\n\n演示用 CLI：读取示例配置，输出缺失变量与建议默认值。\n\n```bash\nconfig-lint --file .env.example\n```\n'),
('01b003proj0000000000000000000003', 'Compose Cookbook',
 '常用 Docker Compose 片段合集：网关、API、静态站一键起。',
 '2026-10-01', 30, '/img/default.png',
 'https://github.com/example/compose-cookbook', 'https://example.com/docs',
 '["01a128b0426f7e2ca83409ec8c567294","01a128b0426f7e32b845844bb54c2a33","01b003demo0000000000000000000003"]',
 '## Compose Cookbook\n\n示例编排说明，对应本仓库 deploy/ 目录的用法摘要。\n\n- gateway（Caddy）\n- api / blog / admin\n- 外置 MySQL 与 Redis\n');
