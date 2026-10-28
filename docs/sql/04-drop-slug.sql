USE go_blog;

-- 去掉实体 slug：文章/项目一律用 uid（UUIDv7 无连字符）访问

-- blog_post.slug
SET @has_post_slug := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'blog_post' AND COLUMN_NAME = 'slug'
);
SET @sql := IF(@has_post_slug > 0, 'ALTER TABLE blog_post DROP COLUMN slug', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- blog_project.uid
SET @has_proj_uid := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'blog_project' AND COLUMN_NAME = 'uid'
);
SET @sql := IF(@has_proj_uid = 0,
  'ALTER TABLE blog_project ADD COLUMN uid CHAR(32) NULL AFTER id',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE blog_project SET uid = LOWER(REPLACE(UUID(), '-', '')) WHERE uid IS NULL OR uid = '';

SET @has_proj_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'blog_project' AND INDEX_NAME = 'uk_blog_project_uid'
);
SET @sql := IF(@has_proj_uk = 0,
  'ALTER TABLE blog_project MODIFY COLUMN uid CHAR(32) NOT NULL, ADD UNIQUE KEY uk_blog_project_uid (uid)',
  'ALTER TABLE blog_project MODIFY COLUMN uid CHAR(32) NOT NULL');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- blog_project.slug
SET @has_proj_slug := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'blog_project' AND COLUMN_NAME = 'slug'
);
SET @sql := IF(@has_proj_slug > 0, 'ALTER TABLE blog_project DROP COLUMN slug', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
