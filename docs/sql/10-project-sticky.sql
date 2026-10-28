-- Author: go-blog — 项目置顶字段（已有行默认不置顶）
USE go_blog;

SET @has_proj_sticky := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'blog_project' AND COLUMN_NAME = 'sticky'
);
SET @sql := IF(
  @has_proj_sticky = 0,
  'ALTER TABLE blog_project ADD COLUMN sticky TINYINT NOT NULL DEFAULT 0 COMMENT ''置顶'' AFTER weight',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
