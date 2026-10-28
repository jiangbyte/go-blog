-- Author: go-blog — 文章/项目草稿状态（已有行默认 published）

SET @has_post_status := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'blog_post' AND COLUMN_NAME = 'status'
);
SET @sql := IF(
  @has_post_status = 0,
  'ALTER TABLE blog_post ADD COLUMN status VARCHAR(32) NOT NULL DEFAULT ''published'' COMMENT ''draft/published'' AFTER published_at, ADD KEY idx_blog_post_status (status)',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_proj_status := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'blog_project' AND COLUMN_NAME = 'status'
);
SET @sql := IF(
  @has_proj_status = 0,
  'ALTER TABLE blog_project ADD COLUMN status VARCHAR(32) NOT NULL DEFAULT ''published'' COMMENT ''draft/published'' AFTER published_at, ADD KEY idx_blog_project_status (status)',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
