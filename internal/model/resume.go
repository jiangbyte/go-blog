package model

import "time"

const (
	ResumeStatusUnpublished = "unpublished"
	ResumeStatusPublished   = "published"
)

// Resume 简历主文档。
type Resume struct {
	ID          string     `gorm:"column:id;size:64;primaryKey"`
	Title       string     `gorm:"column:title;size:128;not null"`
	TemplateID  string     `gorm:"column:template_id;size:64;not null;default:classic"`
	ThemeJSON   string     `gorm:"column:theme_json;type:json;not null"`
	CustomCSS   *string    `gorm:"column:custom_css;type:mediumtext"`
	ProfileJSON string     `gorm:"column:profile_json;type:json;not null"`
	Status      string     `gorm:"column:status;size:32;not null;default:unpublished;index"`
	PublishedAt *time.Time `gorm:"column:published_at"`
	PdfFileID   *string    `gorm:"column:pdf_file_id;size:64"` // 公开 PDF 的 sys_file.id（OSS）
	SortCode    int        `gorm:"column:sort_code;not null;default:0"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (Resume) TableName() string { return "resume" }

// ResumeSection 简历分节。
type ResumeSection struct {
	ID          string    `gorm:"column:id;size:64;primaryKey"`
	ResumeID    string    `gorm:"column:resume_id;size:64;not null;index:idx_resume_section_resume"`
	Type        string    `gorm:"column:type;size:32;not null"`
	Title       string    `gorm:"column:title;size:128;not null"`
	Icon        *string   `gorm:"column:icon;size:128"`
	SortCode    int       `gorm:"column:sort_code;not null;default:0"`
	MetaJSON    *string   `gorm:"column:meta_json;type:json"`
	BodyHTML    *string   `gorm:"column:body_html;type:mediumtext"`
	EntriesJSON *string   `gorm:"column:entries_json;type:json"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ResumeSection) TableName() string { return "resume_section" }
