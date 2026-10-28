USE go_blog;

CREATE TABLE IF NOT EXISTS blog_taxonomy (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  uid CHAR(32) NOT NULL UNIQUE COMMENT 'UUIDv7 without hyphens, public id',
  kind VARCHAR(32) NOT NULL COMMENT 'category | tag | project_tag',
  name VARCHAR(128) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_taxonomy_kind_name (kind, name),
  KEY idx_taxonomy_kind (kind)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
