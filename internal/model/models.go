package model

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/jiangbyte/go-blog/internal/pkg/idgen"
)

type AdminUser struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"size:64;uniqueIndex;not null"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null"`
	Nickname     string    `gorm:"size:64;not null;default:''"`
	Avatar       string    `gorm:"size:512;not null;default:''"`
	Gender       string    `gorm:"size:16;not null;default:UNKNOWN"`
	Signature    string    `gorm:"size:255;not null;default:''"`
	UserType     string    `gorm:"column:user_type;size:16;not null;default:ADMIN"`
	Enabled      int8      `gorm:"not null;default:1"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (AdminUser) TableName() string { return "admin_user" }

type BlogSite struct {
	ID         int8           `gorm:"primaryKey"`
	ConfigJSON datatypes.JSON `gorm:"column:config_json;type:json;not null"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
}

func (BlogSite) TableName() string { return "blog_site" }

type BlogAbout struct {
	ID        int8           `gorm:"primaryKey"`
	DocJSON   datatypes.JSON `gorm:"column:doc_json;type:json;not null"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

func (BlogAbout) TableName() string { return "blog_about" }

// BlogTaxonomy 分类 / 标签 / 项目标签（kind: category | tag | project_tag）
type BlogTaxonomy struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UID       string    `gorm:"column:uid;size:32;uniqueIndex;not null"`
	Kind      string    `gorm:"size:32;not null;index"`
	Name      string    `gorm:"size:128;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (BlogTaxonomy) TableName() string { return "blog_taxonomy" }

func (t *BlogTaxonomy) BeforeCreate(tx *gorm.DB) error {
	if t.UID != "" {
		return nil
	}
	uid, err := idgen.NewUID()
	if err != nil {
		return err
	}
	t.UID = uid
	return nil
}

type BlogPost struct {
	ID                 int64          `gorm:"primaryKey;autoIncrement"`                // 仅内部使用，不对外暴露
	UID                string         `gorm:"column:uid;size:32;uniqueIndex;not null"` // UUIDv7 无连字符，对外主键
	Title              string         `gorm:"size:255;not null"`
	Description        string         `gorm:"type:text"`
	PublishedAt        time.Time      `gorm:"column:published_at;type:date;not null;index"`
	Status             string         `gorm:"column:status;size:32;not null;default:published;index"` // draft | published
	ContentUpdatedAt   *time.Time     `gorm:"column:updated_at;type:date"`
	Cover              string         `gorm:"size:512"`
	CategoriesJSON     datatypes.JSON `gorm:"column:categories_json;type:json;not null"`
	TagsJSON           datatypes.JSON `gorm:"column:tags_json;type:json;not null"`
	Reprint            bool           `gorm:"not null;default:false"`
	Sticky             bool           `gorm:"not null;default:false"`
	HomeVisible        bool           `gorm:"column:home_visible;not null;default:true"`
	ReadingTimeMinutes int            `gorm:"column:reading_time_minutes;not null;default:0"`
	WordCount          int            `gorm:"column:word_count;not null;default:0"`
	Locate             string         `gorm:"size:64"`
	Body               string         `gorm:"type:mediumtext;not null"`
	CreatedAt          time.Time      `gorm:"autoCreateTime"`
}

func (BlogPost) TableName() string { return "blog_post" }

// BeforeCreate 新建文章时自动生成无连字符 UUIDv7。
func (p *BlogPost) BeforeCreate(tx *gorm.DB) error {
	if p.UID != "" {
		return nil
	}
	uid, err := idgen.NewUID()
	if err != nil {
		return err
	}
	p.UID = uid
	return nil
}

type BlogProject struct {
	ID          int64          `gorm:"primaryKey;autoIncrement"`                // 仅内部使用
	UID         string         `gorm:"column:uid;size:32;uniqueIndex;not null"` // UUIDv7 无连字符，对外主键
	Title       string         `gorm:"size:255;not null"`
	Description string         `gorm:"type:text"`
	PublishedAt time.Time      `gorm:"column:published_at;type:date;not null"`
	Status      string         `gorm:"column:status;size:32;not null;default:published;index"` // draft | published
	Progress    string         `gorm:"column:progress;size:32;not null;default:active"`         // planned|active|done|maintained|paused
	Weight      int            `gorm:"not null;default:100;index"`
	Sticky      bool           `gorm:"not null;default:false"`
	Cover       string         `gorm:"size:512"`
	Git         string         `gorm:"size:512"`
	Site        string         `gorm:"size:512"`
	TagsJSON    datatypes.JSON `gorm:"column:tags_json;type:json;not null"`
	Body        string         `gorm:"type:mediumtext;not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
}

func (BlogProject) TableName() string { return "blog_project" }

func (p *BlogProject) BeforeCreate(tx *gorm.DB) error {
	if p.UID != "" {
		return nil
	}
	uid, err := idgen.NewUID()
	if err != nil {
		return err
	}
	p.UID = uid
	return nil
}

func JSONStrings(raw datatypes.JSON) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return []string{}
	}
	return out
}

func JSONMap(raw datatypes.JSON) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func FormatDatePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// BlogStatsTotal 全站累计 PV/UV/文章阅读（由 Redis 异步落库）。
type BlogStatsTotal struct {
	ID        int8      `gorm:"primaryKey"`
	SitePV    int64     `gorm:"column:site_pv;not null;default:0"`
	SiteUV    int64     `gorm:"column:site_uv;not null;default:0"`
	ArticlePV int64     `gorm:"column:article_pv;not null;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (BlogStatsTotal) TableName() string { return "blog_stats_total" }

// BlogStatsDaily 按日聚合（Redis 异步落库）。
type BlogStatsDaily struct {
	Day       time.Time `gorm:"column:day;type:date;primaryKey"`
	SitePV    int64     `gorm:"column:site_pv;not null;default:0"`
	SiteUV    int64     `gorm:"column:site_uv;not null;default:0"`
	ArticlePV int64     `gorm:"column:article_pv;not null;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (BlogStatsDaily) TableName() string { return "blog_stats_daily" }

// BlogEntityViews 文章/项目阅读量（Redis ZSET 异步落库）。
type BlogEntityViews struct {
	Kind      string    `gorm:"column:kind;size:16;primaryKey"` // post | project
	UID       string    `gorm:"column:uid;size:32;primaryKey"`
	ViewCount int64     `gorm:"column:view_count;not null;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (BlogEntityViews) TableName() string { return "blog_entity_views" }

// SysFile 对象存储文件元数据。
type SysFile struct {
	ID              string    `gorm:"column:id;size:64;primaryKey"`
	ObjectName      string    `gorm:"column:object_name;size:255;uniqueIndex;not null"`
	OriginalName    string    `gorm:"column:original_name;size:255;not null"`
	StorageProvider string    `gorm:"column:storage_provider;size:32;not null;index"`
	Bucket          *string   `gorm:"column:bucket;size:255"`
	ContentType     string    `gorm:"column:content_type;size:128;not null"`
	Size            int64     `gorm:"column:size;not null"`
	URL             string    `gorm:"column:url;size:1024;not null"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	CreatedBy       *string   `gorm:"column:created_by;size:64"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
	UpdatedBy       *string   `gorm:"column:updated_by;size:64"`
}

func (SysFile) TableName() string { return "sys_file" }
