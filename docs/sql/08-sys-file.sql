USE go_blog;

-- 文件元数据：对象存储键 + provider，读时按 provider 签发 URL

CREATE TABLE IF NOT EXISTS sys_file (
  id               VARCHAR(64)  NOT NULL COMMENT '主键',
  object_name      VARCHAR(255) NOT NULL COMMENT '对象存储键',
  original_name    VARCHAR(255) NOT NULL COMMENT '原始文件名',
  storage_provider VARCHAR(32)  NOT NULL COMMENT 'minio|qiniu',
  bucket           VARCHAR(255) NULL COMMENT '桶名',
  content_type     VARCHAR(128) NOT NULL COMMENT 'MIME',
  size             BIGINT       NOT NULL COMMENT '字节数',
  url              VARCHAR(1024) NOT NULL COMMENT '访问引用（通常为 object_name）',
  created_at       DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  created_by       VARCHAR(64)  NULL,
  updated_at       DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  updated_by       VARCHAR(64)  NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_sys_file_object_name (object_name),
  KEY idx_sys_file_provider (storage_provider),
  KEY idx_sys_file_created_by (created_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件元数据';
