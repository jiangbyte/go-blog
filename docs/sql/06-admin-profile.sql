USE go_blog;

-- 个人中心：性别 / 签名（已存在列时请跳过）
-- ALTER TABLE admin_user ADD COLUMN gender VARCHAR(16) NOT NULL DEFAULT 'UNKNOWN' AFTER avatar;
-- ALTER TABLE admin_user ADD COLUMN signature VARCHAR(255) NOT NULL DEFAULT '' AFTER gender;

SET @db := DATABASE();
SET @sql := (
  SELECT IF(
    EXISTS(SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@db AND TABLE_NAME='admin_user' AND COLUMN_NAME='gender'),
    'SELECT 1',
    'ALTER TABLE admin_user ADD COLUMN gender VARCHAR(16) NOT NULL DEFAULT ''UNKNOWN'' AFTER avatar'
  )
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := (
  SELECT IF(
    EXISTS(SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=@db AND TABLE_NAME='admin_user' AND COLUMN_NAME='signature'),
    'SELECT 1',
    'ALTER TABLE admin_user ADD COLUMN signature VARCHAR(255) NOT NULL DEFAULT '''' AFTER gender'
  )
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
