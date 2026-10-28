-- Author: go-blog — 项目进度状态（与发布 status 分离；存量默认进行中）

USE go_blog;

SET @has_proj_progress := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'blog_project' AND COLUMN_NAME = 'progress'
);
SET @sql := IF(
  @has_proj_progress = 0,
  'ALTER TABLE blog_project ADD COLUMN progress VARCHAR(32) NOT NULL DEFAULT ''active'' COMMENT ''planned/active/done/maintained/paused'' AFTER status',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
