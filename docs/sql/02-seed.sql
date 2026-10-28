USE go_blog;

-- admin / admin123
INSERT INTO admin_user (id, username, password_hash, nickname, user_type, enabled) VALUES
  (1, 'admin', '$2a$10$liah34hDbqYmH5BSeHEihOviO9FsB4nd56reyGKy9OJnNz.6YS0I.', 'Admin', 'ADMIN', 1)
ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), nickname = VALUES(nickname);

INSERT INTO blog_site (id, config_json) VALUES (1, JSON_OBJECT(
  'title', 'Demo Blog',
  'description', '个人笔记与实践记录 · Demo Blog',
  'site_name', 'Demo Blog',
  'icon', '',
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
    JSON_OBJECT('name', '工具', 'children', JSON_ARRAY(
      JSON_OBJECT('name', 'JSON 格式化', 'href', '/tools/json', 'icon', 'code'),
      JSON_OBJECT('name', '子网划分', 'href', '/tools/subnet', 'icon', 'apps')
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
    'background_url', 'https://t.alcy.cc/fj?u=aside',
    'witty_words', JSON_ARRAY('写清楚，再优化', '先跑通，再抽象'),
    'links', JSON_ARRAY(
      JSON_OBJECT('name', 'GitHub', 'url', 'https://github.com/example', 'icon', 'github'),
      JSON_OBJECT('name', 'Resume', 'url', 'https://resume.example.com', 'icon', 'file')
    )
  ),
  'default_covers', JSON_ARRAY(
    'https://t.alcy.cc/fj?u=d00','https://t.alcy.cc/fj?u=d01','https://t.alcy.cc/fj?u=d02','https://t.alcy.cc/fj?u=d03',
    'https://t.alcy.cc/fj?u=d04','https://t.alcy.cc/fj?u=d05','https://t.alcy.cc/fj?u=d06','https://t.alcy.cc/fj?u=d07'
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
('01a128b0426f7e09932b50bf84e0d1ec', 'category', '软考'),
-- tags
('01a128b0426f7e0bbee94289c21c016b', 'tag', 'Debian'),
('01a128b0426f7e0d921ca123b25d5850', 'tag', '驱动'),
('01a128b0426f7e0fb3a9e4794a7e51a4', 'tag', 'XP-Pen'),
('01a128b0426f7e10a3968788bbed7cf9', 'tag', 'Markdown'),
('01a128b0426f7e12bd141234ab8c88cf', 'tag', '测试'),
('01a128b0426f7e14a8353b52d7ccf082', 'tag', 'Shortcode'),
('01a128b0426f7e1595e68c48c8cd58e2', 'tag', 'Hugo'),
('01a128b0426f7e17a92043a82f8b9673', 'tag', 'Obsidian'),
('01a128b0426f7e19a5592d9ef58f1b26', 'tag', 'Spring'),
('01a128b0426f7e1bbe36faf189bfc193', 'tag', 'AI'),
('01a128b0426f7e1c8bfa76b057483f4e', 'tag', 'RAG'),
('01a128b0426f7e1e8127663b5241bef7', 'tag', '架构师'),
('01a128b0426f7e20877b178c4d7e7dba', 'tag', '软考'),
('01a128b0426f7e22825b66625e34e440', 'tag', 'Docker'),
('01a128b0426f7e24be2156ac079eaf80', 'tag', 'Redis'),
('01a128b0426f7e26a98484f40b16e687', 'tag', 'PostgreSQL'),
-- project_tags
('01a128b0426f7e27b4071d63a7bc1820', 'project_tag', 'Java'),
('01a128b0426f7e29b7d0eac998544287', 'project_tag', '评测'),
('01a128b0426f7e2abcd5009700524e58', 'project_tag', 'Agent'),
('01a128b0426f7e2ca83409ec8c567294', 'project_tag', 'Docker'),
('01a128b0426f7e2eadd749e3f8348586', 'project_tag', 'CLI'),
('01a128b0426f7e3091dce3365ceda7e4', 'project_tag', '前端'),
('01a128b0426f7e32b845844bb54c2a33', 'project_tag', '工具');

DELETE FROM blog_post;
INSERT INTO blog_post (uid, title, description, published_at, updated_at, cover, categories_json, tags_json, reading_time_minutes, word_count, locate, body) VALUES
('01a128a9a640730eaca1c1e0cd6866d6', 'Debian 13 友基 XP-Pen 数位板官方驱动与内核冲突排查固化',
 'Debian 13 + GNOME X11 下 XP-Pen 官方驱动与内核模块冲突的排查步骤与固化方案。',
 '2026-10-07', '2026-10-07', 'https://t.alcy.cc/fj?u=d00',
 '["01a128b0426f7de88e52e2bb7b2ef905"]',
 '["01a128b0426f7e0bbee94289c21c016b","01a128b0426f7e0d921ca123b25d5850","01a128b0426f7e0fb3a9e4794a7e51a4"]',
 12, 2498, 'China',
 '## 现象\n\n数位板偶发失灵。\n\n## 排查\n\n1. 确认 GNOME 会话为 X11\n2. 对比官方驱动与内核版本\n3. 用黑名单固化冲突模块\n'),
('01a128a9a64073408ce8976b098575fd', '01-Markdown特性完整测试',
 '覆盖 CommonMark / GFM / 代码高亮 / Mermaid / KaTeX / 增强 HTML 组件的厨房水槽测试页，用于视觉与构建回归。',
 '2026-10-02', NULL, 'https://t.alcy.cc/fj?u=3ae89bc6',
 '["01a128b0426f7e039bf2ca5b52f27298"]',
 '["01a128b0426f7e10a3968788bbed7cf9","01a128b0426f7e12bd141234ab8c88cf","01a128b0426f7e14a8353b52d7ccf082"]',
 18, 4200, 'China',
 '## Kitchen Sink\n\n完整 Markdown 特性测试正文已迁入仓库后可通过 Admin 更新；此处为种子摘要。\n'),
('01a128a9a6407345acd67fe78a4fa802', 'Hugo相对路径图片显示问题的分析与解决',
 'Obsidian assets 相对路径在 Hugo 中的解析与 render hook 方案。',
 '2026-09-20', '2026-09-21', 'https://t.alcy.cc/fj?u=d01',
 '["01a128b0426f7e06ab33923cdcf5b5f5"]',
 '["01a128b0426f7e1595e68c48c8cd58e2","01a128b0426f7e17a92043a82f8b9673"]',
 8, 1600, 'China',
 '## 问题\n\n相对路径图片在 Hugo 构建后错位。\n\n## 方案\n\n用 render-image hook 解析相对路径。\n'),
('01a128a9a6407349ac39a5eb53fa4bc6', '用 Spring AI 实现 RAG 知识库：对话记忆、向量检索与文档入库',
 '后台多知识库与文件上传，对话侧会话记忆、按库过滤检索与 SSE 流式输出。',
 '2026-08-15', '2026-08-16', 'https://t.alcy.cc/fj?u=d02',
 '["01a128b0426f7e0798c0bd1704698564"]',
 '["01a128b0426f7e19a5592d9ef58f1b26","01a128b0426f7e1bbe36faf189bfc193","01a128b0426f7e1c8bfa76b057483f4e"]',
 15, 3200, 'China',
 '## 概览\n\nSpring AI + 向量库实现可过滤检索的知识库对话。\n');

DELETE FROM blog_project;
INSERT INTO blog_project (uid, title, description, published_at, weight, cover, git, site, tags_json, body) VALUES
('01a128aeed7774bba2e072d1d6f76386', 'ACOJ', '在线评测平台：题目、提交判题与架构风格实践。', '2025-12-01', 10,
 'https://t.alcy.cc/fj?u=p00', 'https://github.com/example', 'https://example.com',
 '["01a128b0426f7e27b4071d63a7bc1820","01a128b0426f7e29b7d0eac998544287"]',
 '## 简介\n\nACOJ 是一套在线评测相关实践项目。\n'),
('01a128aeed7774ed9ddd7556a71d5904', '玄鉴', '终端向的运维/Agent 工具箱，覆盖 Docker 与本机基础设施。', '2026-03-01', 20,
 'https://t.alcy.cc/fj?u=p01', 'https://github.com/example', NULL,
 '["01a128b0426f7e2abcd5009700524e58","01a128b0426f7e2ca83409ec8c567294","01a128b0426f7e2eadd749e3f8348586"]',
 '## 玄鉴\n\n把常用运维动作收敛成可编排的 Agent 工具。\n'),
('01a128aeed7774f5b76258021b391859', 'Startpage', '简洁的浏览器起始页。', '2024-08-01', 30,
 'https://t.alcy.cc/fj?u=p02', NULL, 'https://example.github.io',
 '["01a128b0426f7e3091dce3365ceda7e4","01a128b0426f7e32b845844bb54c2a33"]',
 '轻量起始页，聚合常用链接与搜索。\n');
