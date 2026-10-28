-- Author: go-blog — 公开简历 PDF 存 OSS（sys_file.id）
USE go_blog;
SET NAMES utf8mb4;

SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'resume'
    AND COLUMN_NAME = 'pdf_file_id'
);
SET @sql := IF(
  @col_exists = 0,
  'ALTER TABLE resume ADD COLUMN pdf_file_id VARCHAR(64) NULL COMMENT ''公开 PDF 对应 sys_file.id'' AFTER published_at',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
