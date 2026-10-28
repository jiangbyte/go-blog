USE go_blog;

-- 兼容旧库：若尚无 uid 列则追加并回填；新库（01-schema 已含 uid）跳过 ADD。
SET @has_uid := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'blog_post' AND COLUMN_NAME = 'uid'
);

SET @sql := IF(@has_uid = 0,
  'ALTER TABLE blog_post ADD COLUMN uid CHAR(32) NULL AFTER id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE blog_post SET uid = LOWER(REPLACE(UUID(), '-', '')) WHERE uid IS NULL OR uid = '';

-- 确保 NOT NULL + 唯一索引（已存在则忽略错误由运维手工处理）
SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'blog_post' AND INDEX_NAME = 'uk_blog_post_uid'
);

SET @sql := IF(@has_uk = 0,
  'ALTER TABLE blog_post MODIFY COLUMN uid CHAR(32) NOT NULL, ADD UNIQUE KEY uk_blog_post_uid (uid)',
  'ALTER TABLE blog_post MODIFY COLUMN uid CHAR(32) NOT NULL');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
