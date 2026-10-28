USE go_blog;

CREATE TABLE IF NOT EXISTS admin_user (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(64) NOT NULL DEFAULT '',
  avatar VARCHAR(512) NOT NULL DEFAULT '',
  gender VARCHAR(16) NOT NULL DEFAULT 'UNKNOWN',
  signature VARCHAR(255) NOT NULL DEFAULT '',
  user_type VARCHAR(16) NOT NULL DEFAULT 'ADMIN',
  enabled TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS blog_site (
  id TINYINT PRIMARY KEY,
  config_json JSON NOT NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS blog_about (
  id TINYINT PRIMARY KEY,
  doc_json JSON NOT NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 分类 / 标签 / 项目标签（对外 uid，禁止用名称上 URL）
CREATE TABLE IF NOT EXISTS blog_taxonomy (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  uid CHAR(32) NOT NULL UNIQUE COMMENT 'UUIDv7 without hyphens, public id',
  kind VARCHAR(32) NOT NULL COMMENT 'category | tag | project_tag',
  name VARCHAR(128) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_taxonomy_kind_name (kind, name),
  KEY idx_taxonomy_kind (kind)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS blog_post (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  uid CHAR(32) NOT NULL UNIQUE COMMENT 'UUIDv7 without hyphens, public id',
  title VARCHAR(255) NOT NULL,
  description TEXT,
  published_at DATE NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'published' COMMENT 'draft/published',
  updated_at DATE NULL,
  cover VARCHAR(512) NULL,
  categories_json JSON NOT NULL,
  tags_json JSON NOT NULL,
  reprint TINYINT NOT NULL DEFAULT 0,
  sticky TINYINT NOT NULL DEFAULT 0,
  home_visible TINYINT NOT NULL DEFAULT 1,
  reading_time_minutes INT NOT NULL DEFAULT 0,
  word_count INT NOT NULL DEFAULT 0,
  locate VARCHAR(64) NULL,
  body MEDIUMTEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_post_published (published_at),
  KEY idx_blog_post_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS blog_project (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  uid CHAR(32) NOT NULL UNIQUE COMMENT 'UUIDv7 without hyphens, public id',
  title VARCHAR(255) NOT NULL,
  description TEXT,
  published_at DATE NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'published' COMMENT 'draft/published',
  progress VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT 'planned/active/done/maintained/paused',
  weight INT NOT NULL DEFAULT 100,
  sticky TINYINT NOT NULL DEFAULT 0,
  cover VARCHAR(512) NULL,
  git VARCHAR(512) NULL,
  site VARCHAR(512) NULL,
  tags_json JSON NOT NULL,
  body MEDIUMTEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_project_weight (weight),
  KEY idx_blog_project_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
