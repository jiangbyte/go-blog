-- Author: go-blog — 简历主表 / 分节（无素材库；至多一篇 published）
USE go_blog;
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS resume (
  id VARCHAR(64) NOT NULL COMMENT '主键 UUIDv7',
  title VARCHAR(128) NOT NULL COMMENT '管理端标题',
  template_id VARCHAR(64) NOT NULL DEFAULT 'classic' COMMENT '模板 ID',
  theme_json JSON NOT NULL COMMENT '主题参数',
  custom_css MEDIUMTEXT NULL COMMENT '自定义 CSS（经消毒）',
  profile_json JSON NOT NULL COMMENT '个人资料',
  status VARCHAR(32) NOT NULL DEFAULT 'unpublished' COMMENT 'unpublished|published',
  published_at DATETIME(6) NULL COMMENT '最近发布时间',
  sort_code INT NOT NULL DEFAULT 0 COMMENT '排序',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_resume_status (status),
  KEY idx_resume_updated (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='简历主表';

CREATE TABLE IF NOT EXISTS resume_section (
  id VARCHAR(64) NOT NULL COMMENT '主键',
  resume_id VARCHAR(64) NOT NULL COMMENT '简历 ID',
  type VARCHAR(32) NOT NULL COMMENT 'education|experience|skills|projects|evaluation|custom',
  title VARCHAR(128) NOT NULL COMMENT '分节标题',
  icon VARCHAR(128) NULL COMMENT '图标名',
  sort_code INT NOT NULL DEFAULT 0 COMMENT '排序',
  meta_json JSON NULL COMMENT 'visible 等',
  body_html MEDIUMTEXT NULL COMMENT '正文 HTML（经消毒）',
  entries_json JSON NULL COMMENT '条目列表',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_resume_section_resume (resume_id, sort_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='简历分节';
